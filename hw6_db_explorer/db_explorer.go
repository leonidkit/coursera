package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type DbExplorer struct {
	DB     *sql.DB `json:"-"`
	Tables []Table `json:"tables"`
}

type Column struct {
	Field string
	Type  string

	IsNULL bool
	IsPK   bool
}

type Table struct {
	Name    string
	Columns []Column
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	tables, err := InitTables(db)
	if err != nil {
		return nil, err
	}

	dbex := &DbExplorer{
		db,
		tables,
	}
	return dbex, nil
}

func InitTables(db *sql.DB) ([]Table, error) {
	tables := []Table{}

	trows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer trows.Close()

	var tname string
	for trows.Next() {
		err = trows.Scan(&tname)
		if err != nil {
			return nil, err
		}

		crows, err := db.Query(fmt.Sprintf("SHOW FULL COLUMNS FROM %s", tname))
		if err != nil {
			return nil, err
		}
		defer crows.Close()

		var columns []Column
		var field, typee, isnull, ispk string
		var mock interface{}
		for crows.Next() {
			err = crows.Scan(
				&field,
				&typee,
				&mock,
				&isnull,
				&ispk,
				&mock,
				&mock,
				&mock,
				&mock,
			)
			if err != nil {
				return nil, err
			}
			columns = append(columns, Column{
				field,
				typee,
				isnull == "YES",
				ispk == "PRI",
			})
		}

		tables = append(tables, Table{
			tname,
			columns,
		})
	}
	return tables, nil
}

func (h *DbExplorer) tableExist(name string) *Table {
	for _, t := range h.Tables {
		if t.Name == name {
			return &t
		}
	}
	return nil
}

func columnExist(table *Table, colname string) *Column {
	for _, c := range table.Columns {
		if c.Field == colname {
			return &c
		}
	}
	return nil
}

func parseUrl(url string) (string, string, error) {
	qry := strings.Split(url, "/")
	if len(qry) >= 3 {
		return qry[1], qry[2], nil
	}

	if len(qry) >= 2 {
		return qry[1], "", nil
	}

	return "", "", fmt.Errorf("unknown query")
}

func parseRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	colsTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("unknown error: " + err.Error())
	}

	cols := make([]interface{}, len(colsTypes))
	for i := 0; i < len(cols); i++ {
		cols[i] = new(sql.RawBytes)
	}

	var res []map[string]interface{}
	for rows.Next() {
		rows.Scan(cols...)

		row := make(map[string]interface{}, len(cols))
		for i := 0; i < len(cols); i++ {

			colname := colsTypes[i].Name()
			if *cols[i].(*sql.RawBytes) == nil {
				row[colname] = nil
				continue
			}

			switch colsTypes[i].DatabaseTypeName() {
			case "INT":
				ival, err := strconv.Atoi(string(*cols[i].(*sql.RawBytes)))
				if err != nil {
					return nil, fmt.Errorf("field %s have invalid type", colsTypes[i].Name())
				}
				row[colname] = ival
			case "VARCHAR":
				fallthrough
			case "TEXT":
				row[colname] = string(*cols[i].(*sql.RawBytes))
			default:
				return nil, fmt.Errorf("field %s have invalid type", colsTypes[i].Name())
			}
		}
		res = append(res, row)
	}
	return res, nil
}

func validateValue(val interface{}, col *Column) (interface{}, error) {
	switch val.(type) {
	case float64:
		if col.Type == "float" {
			val = val.(float64)
		} else if col.Type == "int" {
			val = int64(val.(float64))
		} else {
			return nil, fmt.Errorf("field %s have invalid type", col.Field)
		}
	case string:
		if strings.Contains(col.Type, "varchar") || col.Type == "text" {
			val = val.(string)
		} else {
			return nil, fmt.Errorf("field %s have invalid type", col.Field)
		}
	}
	return val, nil
}

func (h *DbExplorer) ReadTables(w http.ResponseWriter, r *http.Request) {
	var tnames []string
	for _, tname := range h.Tables {
		tnames = append(tnames, tname.Name)
	}
	tables := map[string]map[string][]string{
		"response": {
			"tables": tnames,
		},
	}

	err := json.NewEncoder(w).Encode(tables)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *DbExplorer) CreateRow(w http.ResponseWriter, r *http.Request) {
	var rbody map[string]interface{}
	json.NewDecoder(r.Body).Decode(&rbody)
	defer r.Body.Close()

	tname, _, err := parseUrl(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	table := h.tableExist(tname)
	if table == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown table",
		})
		return
	}

	var vals []interface{}
	var valsName string
	var valsPlaceholders string

	for colname, val := range rbody {
		col := columnExist(table, colname)
		if col == nil {
			continue
		}

		if col.IsPK {
			continue
		}

		val, err = validateValue(val, col)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		vals = append(vals, val)
		valsName += colname + ","
		valsPlaceholders += "?,"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tname, valsName[:len(valsName)-1], valsPlaceholders[:len(valsPlaceholders)-1])
	qres, err := h.DB.Exec(query, vals...)
	if err != nil {
		fmt.Println("unknow error: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknow error: " + err.Error(),
		})
		return
	}

	id, err := qres.LastInsertId()
	if err != nil {
		fmt.Println("unknow error: " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknow error: " + err.Error(),
		})
		return
	}

	var pkfield string
	for _, f := range table.Columns {
		if f.IsPK {
			pkfield = f.Field
		}
	}

	created := map[string]map[string]int64{
		"response": {
			pkfield: id,
		},
	}

	err = json.NewEncoder(w).Encode(created)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *DbExplorer) ReadRows(w http.ResponseWriter, r *http.Request) {
	tname, _, err := parseUrl(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	table := h.tableExist(tname)
	if table == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown table",
		})
		return
	}

	vals := r.URL.Query()
	limit := vals.Get("limit")
	offset := vals.Get("offset")
	if offset == "" {
		offset = "0"
	}

	var rows *sql.Rows
	qbuf := fmt.Sprintf("SELECT * FROM %s", tname)
	if limit != "" {
		qbuf += " LIMIT ? OFFSET ?"
		rows, err = h.DB.Query(qbuf, limit, offset)
	} else {
		rows, err = h.DB.Query(qbuf)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown error: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	res, err := parseRows(rows)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	records := map[string]map[string]interface{}{
		"response": {
			"records": res,
		},
	}
	json.NewEncoder(w).Encode(records)
}

func (h *DbExplorer) ReadRow(w http.ResponseWriter, r *http.Request) {
	tname, colid, err := parseUrl(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	table := h.tableExist(tname)
	if table == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown table",
		})
		return
	}

	var pkfield string
	for _, f := range table.Columns {
		if f.IsPK {
			pkfield = f.Field
		}
	}

	qbuf := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", tname, pkfield)
	rows, err := h.DB.Query(qbuf, colid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown error: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	res, err := parseRows(rows)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	if len(res) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "record not found",
		})
		return
	}

	record := map[string]map[string]interface{}{
		"response": {
			"record": res[0],
		},
	}
	json.NewEncoder(w).Encode(record)
}

func (h *DbExplorer) UpdateRow(w http.ResponseWriter, r *http.Request) {
	var rbody map[string]interface{}
	json.NewDecoder(r.Body).Decode(&rbody)
	defer r.Body.Close()

	tname, colid, err := parseUrl(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	table := h.tableExist(tname)
	if table == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown table",
		})
		return
	}

	var vals []interface{}
	var valsPlaceholders string

	for colname, val := range rbody {
		col := columnExist(table, colname)
		if col == nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "unknown column",
			})
			return
		}

		if col.IsPK {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("field %s have invalid type", col.Field),
			})
			return
		}

		if !col.IsNULL && val == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": fmt.Sprintf("field %s have invalid type", col.Field),
			})
			return
		}

		val, err = validateValue(val, col)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		vals = append(vals, val)
		valsPlaceholders += colname + "=?,"
	}
	vals = append(vals, colid)

	var pkfield string
	for _, f := range table.Columns {
		if f.IsPK {
			pkfield = f.Field
			break
		}
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s=?", tname, valsPlaceholders[:len(valsPlaceholders)-1], pkfield)
	qres, err := h.DB.Exec(query, vals...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("unknown error: " + err.Error())
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown error: " + err.Error(),
		})
		return
	}

	raff, err := qres.RowsAffected()
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown error: " + err.Error(),
		})
		return
	}

	record := map[string]map[string]interface{}{
		"response": {
			"updated": raff,
		},
	}
	json.NewEncoder(w).Encode(record)
}

func (h *DbExplorer) DeleteRow(w http.ResponseWriter, r *http.Request) {
	var rbody map[string]interface{}
	json.NewDecoder(r.Body).Decode(&rbody)
	defer r.Body.Close()

	tname, colid, err := parseUrl(r.URL.Path)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	table := h.tableExist(tname)
	if table == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown table",
		})
		return
	}

	var pkfield string
	for _, f := range table.Columns {
		if f.IsPK {
			pkfield = f.Field
		}
	}

	qbuf := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", tname, pkfield)
	qres, err := h.DB.Exec(qbuf, colid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown error: " + err.Error(),
		})
		return
	}

	raff, err := qres.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "unknown error: " + err.Error(),
		})
		return
	}

	record := map[string]map[string]interface{}{
		"response": {
			"deleted": raff,
		},
	}
	json.NewEncoder(w).Encode(record)
}

func (h *DbExplorer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		h.CreateRow(w, r)
	case "GET":
		switch strings.Count(r.URL.Path, "/") {
		case 1:
			if r.URL.Path == "/" {
				h.ReadTables(w, r)
				return
			}
			h.ReadRows(w, r)
		case 2:
			h.ReadRow(w, r)
		}
	case "POST":
		h.UpdateRow(w, r)
	case "DELETE":
		h.DeleteRow(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}
