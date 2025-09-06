package migration

var changes = []Migration{
	// {
	// 	CommitID: "test",
	// 	UpSQL: `
	// 	CREATE TABLE IF NOT EXISTS test_table (
	// 		id BIGINT AUTO_INCREMENT,
	// 		gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	// 		gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	// 		PRIMARY KEY (id)
	// 	) COMMENT='测试表';
	// 	`,
	// 	DownSQL: `
	// 	DROP TABLE IF EXISTS test_table;
	// 	`,
	// },
	// {
	// 	CommitID: "test",
	// 	UpSQL: `
	// 	DROP TABLE IF EXISTS test_table;
	// 	`,
	// 	DownSQL: `
	// 	CREATE TABLE IF NOT EXISTS test_table (
	// 		id BIGINT AUTO_INCREMENT,
	// 		gmt_create TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	// 		gmt_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	// 		PRIMARY KEY (id)
	// 	) COMMENT='测试表';
	// 	`,
	// },
}
