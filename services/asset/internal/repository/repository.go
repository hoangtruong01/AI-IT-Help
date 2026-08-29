package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"eomp/services/asset/internal/model"
)

// Repository interface for Assets and CMDB
type Repository interface {
	// Asset methods
	ListAssets(ctx context.Context, query model.AssetListQuery) (*model.AssetListResponse, error)
	FindAssetByID(ctx context.Context, id string) (*model.Asset, error)
	FindAssetByTag(ctx context.Context, tag string) (*model.Asset, error)
	CreateAsset(ctx context.Context, asset *model.Asset) error
	UpdateAssetStatus(ctx context.Context, id, status, location string, notes *string) error
	AssignAsset(ctx context.Context, id, userID, userName string, deptID *string, condition string, notes *string) error
	ReturnAsset(ctx context.Context, id, condition string, notes *string) error
	GetAssetStats(ctx context.Context) (*model.AssetStatsResponse, error)
	ListAssetAssignments(ctx context.Context, assetID string) ([]model.AssetAssignment, error)
	ListAssignmentsByEmployee(ctx context.Context, userID string) ([]model.EmployeeAssetHistoryItem, error)

	// CMDB methods
	ListCIs(ctx context.Context, env, ciType, status string) ([]model.ConfigurationItem, error)
	FindCIByID(ctx context.Context, id string) (*model.ConfigurationItem, error)
	CreateCI(ctx context.Context, ci *model.ConfigurationItem) error
	UpdateCIStatus(ctx context.Context, id, status string) error
	ListRelationships(ctx context.Context) ([]model.CIRelationship, error)
	CreateRelationship(ctx context.Context, rel *model.CIRelationship) error
	GetTopology(ctx context.Context) (*model.CMDBTopologyGraph, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository constructs a PostgreSQL Asset repository
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListAssets(ctx context.Context, query model.AssetListQuery) (*model.AssetListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	whereClauses := []string{"1=1"}
	args := []any{}
	argIndex := 1

	if query.Category != "" && query.Category != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, query.Category)
		argIndex++
	}

	if query.Status != "" && query.Status != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, query.Status)
		argIndex++
	}

	if query.Location != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("location = $%d", argIndex))
		args = append(args, query.Location)
		argIndex++
	}

	if query.AssignedToUserID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("assigned_to_user_id = $%d", argIndex))
		args = append(args, query.AssignedToUserID)
		argIndex++
	}

	if query.Search != "" {
		pattern := "%" + strings.ToLower(query.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(asset_tag) LIKE $%d OR LOWER(name) LIKE $%d OR LOWER(model) LIKE $%d OR LOWER(serial_number) LIKE $%d)", argIndex, argIndex, argIndex, argIndex))
		args = append(args, pattern)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Total count
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM assets WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count assets: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, asset_tag, name, category, model, serial_number,
			purchase_date::text, purchase_cost, warranty_expiry::text, current_value,
			status, location, assigned_to_user_id, assigned_to_user_name, assigned_at,
			notes, COALESCE(version, 1) AS version, created_at, updated_at
		FROM assets
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query assets: %w", err)
	}
	defer rows.Close()

	assets := []model.Asset{}
	for rows.Next() {
		var a model.Asset
		var modelStr, serialStr, pDate, wExpiry, assignedUID, assignedUName, notesStr sql.NullString
		var assignedAt sql.NullTime

		err := rows.Scan(
			&a.ID, &a.AssetTag, &a.Name, &a.Category, &modelStr, &serialStr,
			&pDate, &a.PurchaseCost, &wExpiry, &a.CurrentValue,
			&a.Status, &a.Location, &assignedUID, &assignedUName, &assignedAt,
			&notesStr, &a.Version, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan asset: %w", err)
		}

		if modelStr.Valid {
			a.Model = &modelStr.String
		}
		if serialStr.Valid {
			a.SerialNumber = &serialStr.String
		}
		if pDate.Valid {
			a.PurchaseDate = &pDate.String
		}
		if wExpiry.Valid {
			a.WarrantyExpiry = &wExpiry.String
		}
		if assignedUID.Valid {
			a.AssignedToUserID = &assignedUID.String
		}
		if assignedUName.Valid {
			a.AssignedToUserName = &assignedUName.String
		}
		if assignedAt.Valid {
			a.AssignedAt = &assignedAt.Time
		}
		if notesStr.Valid {
			a.Notes = &notesStr.String
		}

		assets = append(assets, a)
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))

	return &model.AssetListResponse{
		Data:       assets,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindAssetByID(ctx context.Context, id string) (*model.Asset, error) {
	query := `
		SELECT 
			id, asset_tag, name, category, model, serial_number,
			purchase_date::text, purchase_cost, warranty_expiry::text, current_value,
			status, location, assigned_to_user_id, assigned_to_user_name, assigned_at,
			notes, COALESCE(version, 1) AS version, created_at, updated_at
		FROM assets
		WHERE id = $1
	`
	var a model.Asset
	var modelStr, serialStr, pDate, wExpiry, assignedUID, assignedUName, notesStr sql.NullString
	var assignedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.AssetTag, &a.Name, &a.Category, &modelStr, &serialStr,
		&pDate, &a.PurchaseCost, &wExpiry, &a.CurrentValue,
		&a.Status, &a.Location, &assignedUID, &assignedUName, &assignedAt,
		&notesStr, &a.Version, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query asset by id: %w", err)
	}

	if modelStr.Valid {
		a.Model = &modelStr.String
	}
	if serialStr.Valid {
		a.SerialNumber = &serialStr.String
	}
	if pDate.Valid {
		a.PurchaseDate = &pDate.String
	}
	if wExpiry.Valid {
		a.WarrantyExpiry = &wExpiry.String
	}
	if assignedUID.Valid {
		a.AssignedToUserID = &assignedUID.String
	}
	if assignedUName.Valid {
		a.AssignedToUserName = &assignedUName.String
	}
	if assignedAt.Valid {
		a.AssignedAt = &assignedAt.Time
	}
	if notesStr.Valid {
		a.Notes = &notesStr.String
	}

	return &a, nil
}

func (r *postgresRepository) FindAssetByTag(ctx context.Context, tag string) (*model.Asset, error) {
	query := "SELECT id FROM assets WHERE asset_tag = $1"
	var id string
	err := r.db.QueryRowContext(ctx, query, tag).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return r.FindAssetByID(ctx, id)
}

func (r *postgresRepository) CreateAsset(ctx context.Context, a *model.Asset) error {
	query := `
		INSERT INTO assets (
			asset_tag, name, category, model, serial_number,
			purchase_date, purchase_cost, warranty_expiry, current_value,
			status, location, notes, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, 1, $13, $14
		)
		RETURNING id, version, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		a.AssetTag, a.Name, a.Category, a.Model, a.SerialNumber,
		a.PurchaseDate, a.PurchaseCost, a.WarrantyExpiry, a.CurrentValue,
		a.Status, a.Location, a.Notes, now, now,
	).Scan(&a.ID, &a.Version, &a.CreatedAt, &a.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert asset: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateAssetStatus(ctx context.Context, id, status, location string, notes *string) error {
	query := `
		UPDATE assets
		SET status = $2, location = COALESCE(NULLIF($3, ''), location), notes = COALESCE($4, notes), version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, status, location, notes)
	if err != nil {
		return fmt.Errorf("failed to update asset status: %w", err)
	}
	return nil
}

func (r *postgresRepository) AssignAsset(ctx context.Context, id, userID, userName string, deptID *string, condition string, notes *string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	// 1. Update Asset
	updateQuery := `
		UPDATE assets
		SET status = 'IN_USE', assigned_to_user_id = $2, assigned_to_user_name = $3, assigned_at = $4, version = version + 1, updated_at = $4
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateQuery, id, userID, userName, now); err != nil {
		return fmt.Errorf("failed to update assigned asset: %w", err)
	}

	// 2. Insert Assignment record
	assignQuery := `
		INSERT INTO asset_assignments (asset_id, user_id, user_name, department_id, assigned_at, condition_on_assign, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	if _, err := tx.ExecContext(ctx, assignQuery, id, userID, userName, deptID, now, condition, notes); err != nil {
		return fmt.Errorf("failed to insert asset assignment history: %w", err)
	}

	return tx.Commit()
}

func (r *postgresRepository) ReturnAsset(ctx context.Context, id, condition string, notes *string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	// 1. Update active assignment
	assignQuery := `
		UPDATE asset_assignments
		SET returned_at = $2, condition_on_return = $3, notes = COALESCE($4, notes)
		WHERE asset_id = $1 AND returned_at IS NULL
	`
	if _, err := tx.ExecContext(ctx, assignQuery, id, now, condition, notes); err != nil {
		return fmt.Errorf("failed to close assignment record: %w", err)
	}

	// 2. Update Asset
	updateQuery := `
		UPDATE assets
		SET status = 'IN_STOCK', assigned_to_user_id = NULL, assigned_to_user_name = NULL, assigned_at = NULL, version = version + 1, updated_at = $2
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, updateQuery, id, now); err != nil {
		return fmt.Errorf("failed to return asset to stock: %w", err)
	}

	return tx.Commit()
}

func (r *postgresRepository) GetAssetStats(ctx context.Context) (*model.AssetStatsResponse, error) {
	query := `
		SELECT 
			COUNT(*) AS total_assets,
			COUNT(*) FILTER (WHERE status = 'IN_USE') AS in_use,
			COUNT(*) FILTER (WHERE status = 'IN_STOCK') AS in_stock,
			COUNT(*) FILTER (WHERE status = 'MAINTENANCE') AS in_maintenance,
			COALESCE(SUM(current_value), 0.00) AS total_value
		FROM assets
	`
	var stats model.AssetStatsResponse
	err := r.db.QueryRowContext(ctx, query).Scan(
		&stats.TotalAssets,
		&stats.InUse,
		&stats.InStock,
		&stats.InMaintenance,
		&stats.TotalValue,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset stats: %w", err)
	}
	return &stats, nil
}

func (r *postgresRepository) ListAssetAssignments(ctx context.Context, assetID string) ([]model.AssetAssignment, error) {
	query := `
		SELECT id, asset_id, user_id, user_name, department_id, assigned_at, returned_at, condition_on_assign, condition_on_return, notes
		FROM asset_assignments
		WHERE asset_id = $1
		ORDER BY assigned_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query asset assignments: %w", err)
	}
	defer rows.Close()

	list := []model.AssetAssignment{}
	for rows.Next() {
		var a model.AssetAssignment
		var deptID, condReturn, notesStr sql.NullString
		var returnedAt sql.NullTime

		err := rows.Scan(
			&a.ID, &a.AssetID, &a.UserID, &a.UserName, &deptID,
			&a.AssignedAt, &returnedAt, &a.ConditionOnAssign, &condReturn, &notesStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}

		if deptID.Valid {
			a.DepartmentID = &deptID.String
		}
		if returnedAt.Valid {
			a.ReturnedAt = &returnedAt.Time
		}
		if condReturn.Valid {
			a.ConditionOnReturn = &condReturn.String
		}
		if notesStr.Valid {
			a.Notes = &notesStr.String
		}

		list = append(list, a)
	}
	return list, nil
}

func (r *postgresRepository) ListAssignmentsByEmployee(ctx context.Context, userID string) ([]model.EmployeeAssetHistoryItem, error) {
	query := `
		SELECT 
			aa.id, aa.asset_id, a.asset_tag, a.name, a.category, a.model, a.serial_number, a.status,
			aa.assigned_at, aa.returned_at, aa.condition_on_assign, aa.condition_on_return, aa.notes
		FROM asset_assignments aa
		JOIN assets a ON aa.asset_id = a.id
		WHERE aa.user_id = $1
		ORDER BY aa.assigned_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query employee asset assignments: %w", err)
	}
	defer rows.Close()

	items := []model.EmployeeAssetHistoryItem{}
	for rows.Next() {
		var item model.EmployeeAssetHistoryItem
		var modelStr, serialStr, condReturn, notesStr sql.NullString
		var returnedAt sql.NullTime

		err := rows.Scan(
			&item.AssignmentID, &item.AssetID, &item.AssetTag, &item.AssetName, &item.Category, &modelStr, &serialStr, &item.AssetStatus,
			&item.AssignedAt, &returnedAt, &item.ConditionOnAssign, &condReturn, &notesStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee asset history item: %w", err)
		}

		if modelStr.Valid {
			item.Model = &modelStr.String
		}
		if serialStr.Valid {
			item.SerialNumber = &serialStr.String
		}
		if returnedAt.Valid {
			item.ReturnedAt = &returnedAt.Time
		}
		if condReturn.Valid {
			item.ConditionOnReturn = &condReturn.String
		}
		if notesStr.Valid {
			item.Notes = &notesStr.String
		}

		items = append(items, item)
	}
	return items, nil
}

func (r *postgresRepository) ListCIs(ctx context.Context, env, ciType, status string) ([]model.ConfigurationItem, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1

	if env != "" && env != "All" {
		where = append(where, fmt.Sprintf("environment = $%d", idx))
		args = append(args, env)
		idx++
	}
	if ciType != "" && ciType != "All" {
		where = append(where, fmt.Sprintf("ci_type = $%d", idx))
		args = append(args, ciType)
		idx++
	}
	if status != "" && status != "All" {
		where = append(where, fmt.Sprintf("status = $%d", idx))
		args = append(args, status)
		idx++
	}

	query := fmt.Sprintf(`
		SELECT id, ci_code, name, ci_type, environment, owner_id, owner_name, status, ip_address, asset_id, description, created_at, updated_at
		FROM configuration_items
		WHERE %s
		ORDER BY ci_code ASC
	`, strings.Join(where, " AND "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query CIs: %w", err)
	}
	defer rows.Close()

	cis := []model.ConfigurationItem{}
	for rows.Next() {
		var ci model.ConfigurationItem
		var ownerID, ownerName, ipStr, assetID, descStr sql.NullString
		err := rows.Scan(
			&ci.ID, &ci.CICode, &ci.Name, &ci.CIType, &ci.Environment,
			&ownerID, &ownerName, &ci.Status, &ipStr, &assetID, &descStr,
			&ci.Version, &ci.CreatedAt, &ci.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan CI: %w", err)
		}
		if ownerID.Valid {
			ci.OwnerID = &ownerID.String
		}
		if ownerName.Valid {
			ci.OwnerName = &ownerName.String
		}
		if ipStr.Valid {
			ci.IPAddress = &ipStr.String
		}
		if assetID.Valid {
			ci.AssetID = &assetID.String
		}
		if descStr.Valid {
			ci.Description = &descStr.String
		}

		cis = append(cis, ci)
	}
	return cis, nil
}

func (r *postgresRepository) FindCIByID(ctx context.Context, id string) (*model.ConfigurationItem, error) {
	query := `
		SELECT id, ci_code, name, ci_type, environment, owner_id, owner_name, status, ip_address, asset_id, description, COALESCE(version, 1) AS version, created_at, updated_at
		FROM configuration_items
		WHERE id = $1
	`
	var ci model.ConfigurationItem
	var ownerID, ownerName, ipStr, assetID, descStr sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ci.ID, &ci.CICode, &ci.Name, &ci.CIType, &ci.Environment,
		&ownerID, &ownerName, &ci.Status, &ipStr, &assetID, &descStr,
		&ci.Version, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query CI by id: %w", err)
	}

	if ownerID.Valid {
		ci.OwnerID = &ownerID.String
	}
	if ownerName.Valid {
		ci.OwnerName = &ownerName.String
	}
	if ipStr.Valid {
		ci.IPAddress = &ipStr.String
	}
	if assetID.Valid {
		ci.AssetID = &assetID.String
	}
	if descStr.Valid {
		ci.Description = &descStr.String
	}

	return &ci, nil
}

func (r *postgresRepository) CreateCI(ctx context.Context, ci *model.ConfigurationItem) error {
	query := `
		INSERT INTO configuration_items (ci_code, name, ci_type, environment, owner_id, owner_name, status, ip_address, asset_id, description, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1, $11, $12)
		RETURNING id, version, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		ci.CICode, ci.Name, ci.CIType, ci.Environment,
		ci.OwnerID, ci.OwnerName, ci.Status, ci.IPAddress, ci.AssetID, ci.Description,
		now, now,
	).Scan(&ci.ID, &ci.Version, &ci.CreatedAt, &ci.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert CI: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateCIStatus(ctx context.Context, id, status string) error {
	query := "UPDATE configuration_items SET status = $2, version = version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update CI status: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListRelationships(ctx context.Context) ([]model.CIRelationship, error) {
	query := `
		SELECT 
			r.id, r.parent_ci_id, p.name AS parent_ci_name, p.ci_type AS parent_ci_type,
			r.child_ci_id, c.name AS child_ci_name, c.ci_type AS child_ci_type,
			r.relationship_type, r.impact_weight, r.created_at
		FROM ci_relationships r
		JOIN configuration_items p ON r.parent_ci_id = p.id
		JOIN configuration_items c ON r.child_ci_id = c.id
		ORDER BY r.created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query CI relationships: %w", err)
	}
	defer rows.Close()

	rels := []model.CIRelationship{}
	for rows.Next() {
		var rItem model.CIRelationship
		var pName, pType, cName, cType sql.NullString
		err := rows.Scan(
			&rItem.ID, &rItem.ParentCIID, &pName, &pType,
			&rItem.ChildCIID, &cName, &cType,
			&rItem.RelationshipType, &rItem.ImpactWeight, &rItem.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan relationship: %w", err)
		}
		if pName.Valid {
			rItem.ParentCIName = &pName.String
		}
		if pType.Valid {
			rItem.ParentCIType = &pType.String
		}
		if cName.Valid {
			rItem.ChildCIName = &cName.String
		}
		if cType.Valid {
			rItem.ChildCIType = &cType.String
		}

		rels = append(rels, rItem)
	}
	return rels, nil
}

func (r *postgresRepository) CreateRelationship(ctx context.Context, rel *model.CIRelationship) error {
	query := `
		INSERT INTO ci_relationships (parent_ci_id, child_ci_id, relationship_type, impact_weight, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(ctx, query, rel.ParentCIID, rel.ChildCIID, rel.RelationshipType, rel.ImpactWeight, now).Scan(&rel.ID, &rel.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert relationship: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetTopology(ctx context.Context) (*model.CMDBTopologyGraph, error) {
	nodes, err := r.ListCIs(ctx, "", "", "")
	if err != nil {
		return nil, err
	}
	edges, err := r.ListRelationships(ctx)
	if err != nil {
		return nil, err
	}
	return &model.CMDBTopologyGraph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}
