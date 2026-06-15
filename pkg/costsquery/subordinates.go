package costsquery

import (
	"fmt"

	sq "github.com/n-r-w/squirrel"

	"github.com/SOTBI-LLC/sotbi.lib/pkg/commonqueries"
	"github.com/SOTBI-LLC/sotbi.lib/pkg/filtering/squirrel_fltering"
	"github.com/SOTBI-LLC/sotbi.lib/pkg/utils"
)

// BuildSubordinatesCostsSQL builds the subordinates costs query used by excelgen and sotbisrv.
func BuildSubordinatesCostsSQL(prm utils.CostsRequestParams, filterModel string) (sql string, args []any, err error) {
	periodStart := prm.Start.Format("2006-01-02")
	periodEnd := prm.End.Format("2006-01-02")

	userTotals := buildSubordinatesPeriodTotals(periodStart, periodEnd)

	query := sq.
		Select(
			"cr.id",
			`users."user" as user_name`,
			`COALESCE(CASE WHEN (u1."user" IS NOT NULL) THEN CONCAT(u."user", ',', u1."user") ELSE u."user" END, '') AS debtor_group`, //nolint:lll
			"cr.date",
			"to_char(cr.date, 'TMDay') as day_of_week",
			"cr.user_id",
			"cr.debtor_id",
			"((cr.minutes_costs::text) || 'minutes')::interval as time",
			"COALESCE(cr.description, '') AS description",
			"cr.minutes_costs",
			"pd.name AS debtor_name",
			"pd.project_name",
			"cr.work_category_id",
			"wc.name as work_category_name",
			"COALESCE(us.unit1, '') as unit1",
			"COALESCE(us.unit2, '') as unit2",
			"COALESCE(user_totals.total_minutes, 0) AS total_minutes",
			"COALESCE(ROUND(cr.minutes_costs::numeric / NULLIF(COALESCE(user_totals.total_minutes, 0), 0) * 100, 3), 0) AS minutes_share", //nolint:lll
			"COALESCE(pos.name, '') AS position",
		).
		From("costs_real cr").
		Join("users on cr.user_id = users.id").
		LeftJoin("work_categories wc ON wc.id = cr.work_category_id").
		LeftJoin("projects_debtors pd ON cr.debtor_id = pd.id").
		LeftJoin("plist ON pd.project_id = plist.id").
		LeftJoin("staffs s ON plist.group_id = s.id").
		LeftJoin("users u ON s.user_id = u.id").
		LeftJoin("staffs s1 ON plist.manager_id = s1.id").
		LeftJoin("users u1 ON s1.user_id = u1.id").
		LeftJoin(commonqueries.NamedTable(commonqueries.UsersStaffsLight, "us") + " ON cr.user_id = us.id").
		LeftJoin("staffs s3 ON s3.user_id = cr.user_id").
		JoinClause(userTotals.Prefix("LEFT JOIN (").Suffix(") user_totals ON user_totals.user_id = cr.user_id")).
		LeftJoin(`LATERAL (
    SELECT p.name
    FROM users_positions up
    JOIN positions p ON p.id = up.position_id
    WHERE up.user_id = cr.user_id AND up.date <= cr.date
    ORDER BY up.date DESC
    LIMIT 1
) pos ON true`).
		OrderBy("date").
		PlaceholderFormat(sq.Dollar)

	query = applySubordinatesCostsScope(query, prm, "cr")
	query = applySubordinatesPeriodFilter(query, periodStart, periodEnd, "cr")
	query = squirrel_fltering.CreateFilter(query, filterModel, "cr")

	sql, args, err = query.ToSql()
	if err != nil {
		return "", nil, fmt.Errorf("build sql: %w", err)
	}

	return sql, args, nil
}

func buildSubordinatesPeriodTotals(periodStart, periodEnd string) sq.SelectBuilder {
	return sq.Select(
		"c2.user_id",
		"COALESCE(SUM(c2.minutes_costs), 0) AS total_minutes",
	).
		From("costs_real c2").
		Where(sq.And{
			sq.Expr("c2.date::date >= ?::date", periodStart),
			sq.Expr("c2.date::date <= ?::date", periodEnd),
		}).
		GroupBy("c2.user_id").
		PlaceholderFormat(sq.Dollar)
}

func applySubordinatesPeriodFilter(
	query sq.SelectBuilder,
	periodStart, periodEnd, alias string,
) sq.SelectBuilder {
	return query.Where(sq.And{
		sq.Expr(alias+".date::date >= ?::date", periodStart),
		sq.Expr(alias+".date::date <= ?::date", periodEnd),
	})
}

func applySubordinatesCostsScope(
	query sq.SelectBuilder,
	prm utils.CostsRequestParams,
	alias string,
) sq.SelectBuilder {
	if len(prm.Debtors) > 0 {
		return query.Where(sq.Or{
			sq.Eq{alias + ".user_id": prm.Users},
			sq.Eq{alias + ".debtor_id": prm.Debtors},
		})
	}

	return query.Where(sq.Eq{alias + ".user_id": prm.Users})
}
