package commonqueries

import "fmt"

func NamedTable(selectQuery, tableName string) string {
	return fmt.Sprintf(`(%s) AS %q`, selectQuery, tableName)
}

const (
	SelectOldRemainings = `
SELECT exchange_files.id
     , exchange_files.start_date
     , exchange_files.end_date
     , CASE
           WHEN remainings.account IS NULL THEN exchange_files.account
           ELSE remainings.account
    END                                        AS account
     , (SELECT r1.initial_balance
        FROM remainings r1
        WHERE r1.exchange_file_id = exchange_files.id
          AND r1.start_date = exchange_files.start_date
          AND r1.account::text = remainings.account::text
        GROUP BY r1.initial_balance
        HAVING min(r1.start_date) IS NOT NULL) AS initial_balance
     , sum(remainings.income)                  AS income
     , sum(remainings.write_off)               AS write_off
     , (SELECT r2.final_balance
        FROM remainings r2
        WHERE r2.exchange_file_id = exchange_files.id
          AND r2.end_date = exchange_files.end_date
          AND r2.account::text = remainings.account::text
        GROUP BY r2.final_balance
        HAVING max(r2.end_date) IS NOT NULL)   AS final_balance
     , exchange_files.file
     , exchange_files.type
     , exchange_files.creator_id
     , exchange_files.created_at
FROM exchange_files
         LEFT JOIN remainings ON remainings.exchange_file_id = exchange_files.id
GROUP BY exchange_files.id, remainings.account
`
	FileLinks = `
SELECT payment_attachments.file
     , payment_attachments.original_file_name
FROM payment_attachments
WHERE payment_attachments.file IS NOT NULL
UNION
SELECT defrayments.file
     , defrayments.original_file_name
FROM defrayments
WHERE defrayments.file IS NOT NULL
UNION
SELECT insurance_attachments.file
     , insurance_attachments.original_file_name
FROM insurance_attachments
WHERE insurance_attachments.file IS NOT NULL
UNION
SELECT egrn_attachments.file
     , egrn_attachments.original_file_name
FROM egrn_attachments
WHERE egrn_attachments.file IS NOT NULL
UNION
SELECT real_estates.file
     , real_estates.original_file_name
FROM real_estates
WHERE real_estates.file IS NOT NULL
UNION
SELECT attachments.file
     , attachments.original_file_name
FROM attachments
WHERE attachments.file IS NOT NULL
UNION
SELECT calculations.file
     , calculations.original_file_name
FROM calculations
WHERE length(calculations.file::text) > 0`
	EgrnRequestsByObject = `
SELECT e.id
     , e.debtor_id
     , e.project_id
     , p.name      AS project_name
     , d.name      AS debtor_name
     , e.status
     , e.statement_type
     , e.providing_way
     , e.description
     , e.doer_comment
     , r.cadastral_no
     , r.parameters
     , e.created_at
     , u1."user"   AS created_by_name
     , e.updated_at
     , e.passport
     , r.request_num
     , count(r.id) AS real_estates_count
     , CASE
           WHEN e.statement_type = 'person'::egrn_requests_statement_type THEN
               CASE
                   WHEN e.rightholder = 'bankruptcy'::egrn_requests_rightholder THEN d.name
                   ELSE coalesce(e.thirdperson_inn, e.fio)
                   END
           ELSE
               CASE
                   WHEN e.on_behalf_of = 'bankruptcy'::egrn_requests_on_behalf_of THEN
                       CASE
                           WHEN e.rightholder = 'bankruptcy'::egrn_requests_rightholder THEN d.name
                           ELSE '3-е лицо'::character varying
                           END
                   ELSE '3-е лицо'::character varying
                   END
    END            AS rightholder_name
FROM egrn_requests e
         JOIN plist p ON p.id = e.project_id
         LEFT JOIN debtors d ON d.id = e.debtor_id
         LEFT JOIN real_estates r ON r.egrn_request_id = e.id
         JOIN egrn_request_histories eh ON eh.id = e.id AND lower(eh.action::text) = 'insert'::text
         LEFT JOIN users u1 ON u1.id = eh.updated_by_id
GROUP BY e.id, d.name, p.name, r.cadastral_no, r.parameters, r.request_num, u1."user"
ORDER BY e.created_at DESC`
	EgrnRequestsByClaim = `
SELECT e.id
     , e.debtor_id
     , e.project_id
     , e.status
     , p.name      AS project_name
     , d.name      AS debtor_name
     , e.statement_type
     , e.rightholder
     , e.providing_way
     , e.description
     , e.doer_comment
     , e.created_at
     , u1.id       AS created_by_id
     , u1."user"   AS created_by_name
     , e.updated_at
     , u.id        AS updated_by_id
     , u."user"    AS updated_by_name
     , count(r.id) AS real_estates_count
     , CASE
           WHEN e.statement_type = 'person'::egrn_requests_statement_type THEN
               CASE
                   WHEN e.rightholder = 'bankruptcy'::egrn_requests_rightholder THEN d.name
                   ELSE coalesce(e.thirdperson_inn, e.fio)
                   END
           ELSE
               CASE
                   WHEN e.on_behalf_of = 'bankruptcy'::egrn_requests_on_behalf_of THEN
                       CASE
                           WHEN e.rightholder = 'bankruptcy'::egrn_requests_rightholder THEN d.name
                           ELSE '3-е лицо'::character varying
                           END
                   ELSE '3-е лицо'::character varying
                   END
    END            AS rightholder_name
FROM egrn_requests e
         JOIN plist p ON p.id = e.project_id
         LEFT JOIN debtors d ON d.id = e.debtor_id
         LEFT JOIN real_estates r ON e.id = r.egrn_request_id
         LEFT JOIN users u ON u.id = e.updated_by_id
         JOIN egrn_request_histories eh ON eh.id = e.id AND lower(eh.action::text) = 'insert'::text
         LEFT JOIN users u1 ON u1.id = eh.updated_by_id
GROUP BY e.id, d.name, p.name, u1.id, u.id
ORDER BY e.created_at DESC`
	ActivePeriodsUsers = `
SELECT active_periods.month                                                                     AS month
     , active_periods.first_day_month                                                            AS first_day_month
     , active_periods.holidays                                                                   AS holidays
     , active_periods.editable                                                                   AS editable
     , (SELECT count(*) AS count
        FROM get_completed_users(active_periods.first_day_month)
          get_completed_users(id, "user", email, data, hire, quit, tel, group_id, settings))     AS completed_users
     , (SELECT count(*) AS count
        FROM get_not_completed_users(active_periods.first_day_month)
          get_not_completed_users(id, "user", email, data, hire, quit, tel, group_id, settings)) AS not_completed_users
     , (SELECT count(*) AS count
        FROM get_projects(active_periods.first_day_month)
          get_projects(project_id))                                                              AS projects
FROM active_periods
WHERE active_periods.first_day_month >= (last_day(current_date) + '1 day'::interval - '1 year'::interval)::date
ORDER BY active_periods.first_day_month DESC
`
	UsersStaffsLight = `
SELECT tree.id,
    tree."user",
    tree.unit1_id,
    tree.unit2_id,
    s1.name AS unit1,
    s2.name AS unit2
FROM (
         SELECT u.id,
             u."user",
             cte.unit1_id,
             cte.unit2_id
         FROM users u
                  LEFT JOIN LATERAL (
             WITH RECURSIVE cte(id, name, parent_id, type, user_id) AS (
                 SELECT staffs.id, staffs.name, staffs.parent_id, staffs.type, staffs.user_id
                 FROM staffs
                 WHERE staffs.user_id = u.id
                 UNION ALL
                 SELECT p.id, p.name, p.parent_id, p.type, p.user_id
                 FROM staffs p
                          JOIN cte cte_1 ON p.id = cte_1.parent_id
             )
             SELECT
                 min(CASE WHEN cte.type = 5 THEN cte.id END) AS unit1_id,
                 min(CASE WHEN cte.type = 4 THEN cte.id END) AS unit2_id
             FROM cte
             ) cte ON TRUE
         ORDER BY u."user"
     ) tree
         LEFT JOIN staffs s1 ON s1.id = tree.unit1_id
         LEFT JOIN staffs s2 ON s2.id = tree.unit2_id
`
	UsersStaffs = `
SELECT tree.id
     , tree."user"
     , tree.avatar
     , tree.data
     , tree.pr
     , tree.uh
     , tree.tel
     , tree.group_id
     , tree.email
     , tree.settings
     , tree.updated_at
     , tree.updated_by
     , tree.unit1_id
     , tree.unit2_id
     , s1.name AS unit1
     , s2.name AS unit2
FROM (SELECT u.id
           , u."user"
           , u.avatar
           , u.data
           , u.pr
           , u.uh
           , u.tel
           , u.group_id
           , u.email
           , u.settings
           , u.updated_at
           , u.updated_by
           , cte.unit1_id
           , cte.unit2_id
           FROM users u
                  LEFT JOIN LATERAL (
             WITH RECURSIVE cte(id, name, parent_id, type, user_id) AS (
                 SELECT staffs.id, staffs.name, staffs.parent_id, staffs.type, staffs.user_id
                 FROM staffs
                 WHERE staffs.user_id = u.id
                 UNION ALL
                 SELECT p.id, p.name, p.parent_id, p.type, p.user_id
                 FROM staffs p
                          JOIN cte cte_1 ON p.id = cte_1.parent_id
             )
             SELECT
                 MIN(CASE WHEN cte.type = 5 THEN cte.id END) AS unit1_id,
                 MIN(CASE WHEN cte.type = 4 THEN cte.id END) AS unit2_id
             FROM cte
             ) cte ON true
      ORDER BY u."user") tree
         LEFT JOIN staffs s1 ON s1.id = tree.unit1_id
         LEFT JOIN staffs s2 ON s2.id = tree.unit2_id`
)
