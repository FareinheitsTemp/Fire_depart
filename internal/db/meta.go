package db

// Column — опис колонки таблиці для ERD-мапи та UI.
type Column struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Primary  bool   `json:"primary"`
}

// ForeignKey — зовнішній ключ таблиці.
type ForeignKey struct {
	Column    string `json:"column"`
	RefTable  string `json:"refTable"`
	RefColumn string `json:"refColumn"`
}

// Table — метадані таблиці.
type Table struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	ForeignKeys []ForeignKey `json:"foreignKeys"`
}

func pk(name, typ string) Column {
	return Column{Name: name, Type: typ, Primary: true}
}

func col(name, typ string, nullable bool) Column {
	return Column{Name: name, Type: typ, Nullable: nullable}
}

func fk(column, refTable string) ForeignKey {
	return ForeignKey{Column: column, RefTable: refTable, RefColumn: "id"}
}

// schemaMeta — статичні метадані схеми (пізніше читатимемо з pg_catalog живої БД).
var schemaMeta = []Table{
	{
		Name: "fire_stations",
		Columns: []Column{
			pk("id", "integer"),
			col("name", "varchar(120)", false),
			col("address", "varchar(200)", false),
			col("phone", "varchar(30)", false),
		},
	},
	{
		Name: "ranks",
		Columns: []Column{
			pk("id", "integer"),
			col("title", "varchar(60)", false),
		},
	},
	{
		Name: "employees",
		Columns: []Column{
			pk("id", "integer"),
			col("station_id", "integer", false),
			col("rank_id", "integer", true),
			col("full_name", "varchar(120)", false),
			col("phone", "varchar(30)", false),
			col("hired_at", "date", true),
			col("is_active", "boolean", false),
		},
		ForeignKeys: []ForeignKey{
			fk("station_id", "fire_stations"),
			fk("rank_id", "ranks"),
		},
	},
	{
		Name: "shifts",
		Columns: []Column{
			pk("id", "integer"),
			col("station_id", "integer", false),
			col("starts_at", "timestamp", false),
			col("ends_at", "timestamp", false),
		},
		ForeignKeys: []ForeignKey{fk("station_id", "fire_stations")},
	},
	{
		Name: "shift_assignments",
		Columns: []Column{
			pk("id", "integer"),
			col("shift_id", "integer", false),
			col("employee_id", "integer", false),
		},
		ForeignKeys: []ForeignKey{
			fk("shift_id", "shifts"),
			fk("employee_id", "employees"),
		},
	},
	{
		Name: "vehicles",
		Columns: []Column{
			pk("id", "integer"),
			col("station_id", "integer", false),
			col("call_sign", "varchar(20)", false),
			col("vehicle_type", "varchar(60)", false),
			col("status", "varchar(20)", false),
			col("purchase_date", "date", true),
		},
		ForeignKeys: []ForeignKey{fk("station_id", "fire_stations")},
	},
	{
		Name: "incident_types",
		Columns: []Column{
			pk("id", "integer"),
			col("title", "varchar(80)", false),
		},
	},
	{
		Name: "incidents",
		Columns: []Column{
			pk("id", "integer"),
			col("station_id", "integer", false),
			col("type_id", "integer", false),
			col("address", "varchar(200)", false),
			col("description", "text", false),
			col("received_at", "timestamp", false),
			col("closed_at", "timestamp", true),
			col("status", "varchar(20)", false),
		},
		ForeignKeys: []ForeignKey{
			fk("station_id", "fire_stations"),
			fk("type_id", "incident_types"),
		},
	},
	{
		Name: "incident_vehicles",
		Columns: []Column{
			pk("id", "integer"),
			col("incident_id", "integer", false),
			col("vehicle_id", "integer", false),
			col("dispatched_at", "timestamp", false),
		},
		ForeignKeys: []ForeignKey{
			fk("incident_id", "incidents"),
			fk("vehicle_id", "vehicles"),
		},
	},
	{
		Name: "incident_responders",
		Columns: []Column{
			pk("id", "integer"),
			col("incident_id", "integer", false),
			col("employee_id", "integer", false),
		},
		ForeignKeys: []ForeignKey{
			fk("incident_id", "incidents"),
			fk("employee_id", "employees"),
		},
	},
	{
		Name: "incident_reports",
		Columns: []Column{
			pk("id", "integer"),
			col("incident_id", "integer", false),
			col("summary", "text", false),
			col("damage_uah", "numeric(14,2)", false),
			col("casualties", "integer", false),
			col("created_at", "timestamp", false),
		},
		ForeignKeys: []ForeignKey{fk("incident_id", "incidents")},
	},
	{
		Name: "schema_layouts",
		Columns: []Column{
			pk("id", "integer"),
			col("view_name", "varchar(60)", false),
			col("positions", "jsonb", false),
			col("updated_at", "timestamp", false),
		},
	},
}

// SchemaMeta повертає метадані схеми для ERD-мапи.
func SchemaMeta() []Table {
	return schemaMeta
}
