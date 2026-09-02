package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"eomp/services/notification/internal/model"

	_ "github.com/lib/pq"
)

func getNotificationIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	required := os.Getenv("INTEGRATION_REQUIRED") != ""
	dsn := os.Getenv("NOTIFICATION_INTEGRATION_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		if required {
			t.Fatal("NOTIFICATION_INTEGRATION_DSN is required")
		}
		t.Skip("skipping notification PostgreSQL integration test (NOTIFICATION_INTEGRATION_DSN not set)")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		if required {
			t.Fatalf("open notification PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot open notification PostgreSQL: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if required {
			t.Fatalf("ping notification PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot ping notification PostgreSQL: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestNotificationIntegration_ReadReceiptsAreRecipientScoped(t *testing.T) {
	db := getNotificationIntegrationDB(t)
	repo := NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	suffix := fmt.Sprintf("%d", time.Now().UnixNano())
	aliceID := "gate-d-alice-" + suffix
	bobID := "gate-d-bob-" + suffix
	role := "ROLE_GATE_D_" + suffix
	category := "GATE_D_" + suffix

	makeNotification := func(recipientID, title string) *model.Notification {
		return &model.Notification{
			RecipientID:    recipientID,
			RecipientEmail: recipientID + "@local.test",
			Title:          title,
			Message:        "recipient-scoped read receipt integration test",
			Category:       category,
			Priority:       model.PriorityMedium,
			Channel:        model.ChannelInApp,
		}
	}

	direct := makeNotification(aliceID, "Direct notification")
	roleWide := makeNotification(role, "Role notification")
	broadcast := makeNotification("all", "Broadcast notification")
	for _, notification := range []*model.Notification{direct, roleWide, broadcast} {
		if err := repo.CreateNotification(ctx, notification); err != nil {
			t.Fatalf("create notification %q: %v", notification.Title, err)
		}
	}
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM notifications WHERE id = ANY($1::uuid[])`,
			fmt.Sprintf("{%s,%s,%s}", direct.ID, roleWide.ID, broadcast.ID))
	})

	query := func(recipientID string) *model.NotificationListResponse {
		result, err := repo.ListNotifications(ctx, model.NotificationListQuery{
			RecipientID: recipientID, RecipientRole: role, Category: category, Page: 1, PageSize: 20,
		})
		if err != nil {
			t.Fatalf("list notifications for %s: %v", recipientID, err)
		}
		return result
	}

	if initial := query(aliceID); initial.Total != 3 || initial.UnreadCount != 3 {
		t.Fatalf("expected Alice to start with 3 unread notifications, got total=%d unread=%d", initial.Total, initial.UnreadCount)
	}
	if err := repo.MarkAsRead(ctx, roleWide.ID, aliceID, role); err != nil {
		t.Fatalf("mark Alice role notification read: %v", err)
	}
	if alice := query(aliceID); alice.UnreadCount != 2 {
		t.Fatalf("expected Alice to have 2 unread after one receipt, got %d", alice.UnreadCount)
	}
	if bob := query(bobID); bob.Total != 2 || bob.UnreadCount != 2 {
		t.Fatalf("Alice receipt leaked to Bob: total=%d unread=%d", bob.Total, bob.UnreadCount)
	}
	if err := repo.MarkAsRead(ctx, direct.ID, bobID, role); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected unauthorized direct receipt to return sql.ErrNoRows, got %v", err)
	}
	if err := repo.MarkAllAsRead(ctx, bobID, role); err != nil {
		t.Fatalf("mark Bob role/broadcast notifications read: %v", err)
	}
	if bob := query(bobID); bob.UnreadCount != 0 {
		t.Fatalf("expected Bob's two visible notifications to be read, got unread=%d", bob.UnreadCount)
	}
	if alice := query(aliceID); alice.UnreadCount != 2 {
		t.Fatalf("Bob receipts changed Alice state: unread=%d", alice.UnreadCount)
	}
}
