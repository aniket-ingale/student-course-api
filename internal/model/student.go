// Package model holds the domain types persisted by the service.
package model

// Student maps to the students table. student_id is a DB-assigned identity
// column; autoIncrement tells GORM never to write it on insert.
type Student struct {
	StudentID int    `gorm:"column:student_id;primaryKey;autoIncrement" json:"studentId"`
	Name      string `gorm:"column:name"                                json:"name"`
	Address   string `gorm:"column:address"                             json:"address"`
	Grade     int    `gorm:"column:grade"                               json:"grade"`
}

// TableName pins the table name so GORM does not rely on pluralization rules.
func (Student) TableName() string { return "students" }
