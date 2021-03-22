package utils

import (
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
)

// TableFilter is auxiliary struct to contain method of filtering
type TableFilter struct {
	Value interface{} `form:"value" json:"value" binding:"required"`
	Field string      `form:"field" json:"field" binding:"required"`
}

// TableSort is auxiliary struct to contain method of sorting
type TableSort struct {
	Prop  string `form:"prop" json:"prop" binding:"omitempty"`
	Order string `form:"order" json:"order" binding:"omitempty"`
}

// TableQuery is main struct to contain input params
type TableQuery struct {
	// Number of page (since 1)
	Page int `form:"page" json:"page" binding:"min=1,required" default:"1" minimum:"1"`
	// Amount items per page (min 5, max 100)
	Size int `form:"pageSize" json:"pageSize" binding:"min=5,max=100,required" default:"5" minimum:"5" maximum:"100"`
	// Type of request
	Type string `form:"type" json:"type" binding:"oneof=sort filter init page size,required" default:"init" enums:"sort,filter,init,page,size"`
	// Language of result data
	Lang string `form:"lang" json:"lang" binding:"oneof=ru en,required" default:"en" enums:"en,ru"`
	// Sorting result on server e.g. {"prop":"...","order":"..."}
	//   field order is "ascending" or "descending" value
	Sort TableSort `form:"sort" json:"sort" binding:"required" swaggertype:"string" default:"{}"`
	// Filtering result on server e.g. {"value":[...],"field":"..."}
	//   field value should be integer or string or array type
	Filters []TableFilter `form:"filters[]" json:"filters[]" binding:"omitempty" swaggertype:"array,string"`
	// non input arguments
	table      string            `form:"-" json:"-"`
	sqlMappers map[string]string `form:"-" json:"-"`
}

// Init is function to set table name and sql mapping to data columns
func (q *TableQuery) Init(table string, sqlMappers map[string]string) {
	q.table = table
	q.sqlMappers = make(map[string]string, 0)
	for k, v := range sqlMappers {
		v = strings.ReplaceAll(v, "{{lang}}", q.Lang)
		v = strings.ReplaceAll(v, "{{type}}", q.Type)
		v = strings.ReplaceAll(v, "{{table}}", q.table)
		v = strings.ReplaceAll(v, "{{page}}", strconv.Itoa(q.Page))
		v = strings.ReplaceAll(v, "{{size}}", strconv.Itoa(q.Size))
		q.sqlMappers[k] = v
	}
}

// Ordering is function to get order of data rows according with input params
func (q *TableQuery) Ordering() func(db *gorm.DB) *gorm.DB {
	field := ""
	arrow := "DESC"
	if q.Sort.Order == "ascending" {
		arrow = "ASC"
	}
	if v, ok := q.sqlMappers[q.Sort.Prop]; ok {
		field = v
	}
	return func(db *gorm.DB) *gorm.DB {
		if field == "" {
			return db
		}
		return db.Order(field + " " + arrow)
	}
}

// Paginate is function to navigate between pages according with input params
func (q *TableQuery) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if q.Page <= 0 && q.Size > 0 {
			return db.Limit(q.Size)
		} else if q.Page > 0 && q.Size > 0 {
			offset := (q.Page - 1) * q.Size
			return db.Offset(offset).Limit(q.Size)
		}
		return db
	}
}

// DataFilter is function to build main data filter from filters input params
func (q *TableQuery) DataFilter() func(db *gorm.DB) *gorm.DB {
	fs := make(map[string]string, 0)
	fl := make(map[string][]string, 0)
	for _, f := range q.Filters {
		if _, ok := q.sqlMappers[f.Field]; ok {
			if v, ok := f.Value.(string); ok && v != "" {
				fs[f.Field] = "%" + Escape(v) + "%"
			}
			if v, ok := f.Value.([]interface{}); ok && len(v) != 0 {
				var vs []string
				for _, ti := range v {
					if ts, ok := ti.(string); ok {
						vs = append(vs, ts)
					}
				}
				if l, ok := fl[f.Field]; ok {
					fl[f.Field] = append(l, vs...)
				} else {
					fl[f.Field] = vs
				}
			}
		}
	}
	return func(db *gorm.DB) *gorm.DB {
		for k, v := range fs {
			db = db.Where(q.sqlMappers[k]+" LIKE ?", v)
		}
		for k, v := range fl {
			db = db.Where(q.sqlMappers[k]+" IN (?)", v)
		}
		return db
	}
}

// Query is function to retrieve table data according with input params
func (q *TableQuery) Query(db *gorm.DB, result interface{},
	funcs ...func(*gorm.DB) *gorm.DB) (uint64, error) {
	var total uint64
	err := db.Scopes(funcs...).Scopes(q.DataFilter()).Table(q.table).
		Count(&total).Scopes(q.Ordering(), q.Paginate()).Find(result).Error
	return total, err
}
