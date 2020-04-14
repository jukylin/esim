package entity

import (

 "github.com/jinzhu/gorm"
 "database/sql"
 "time"
)


type Test struct{

//user name
UserName sql.NullString `gorm:"column:user_name"`

Id int `gorm:"column:id;primary_key"`

UpdateTime time.Time `gorm:"column:update_time"`

}

// delete field
func (c *Test) DelKey() string {
	return ""
}


//自动增加时间
func (this *Test) BeforeCreate(scope *gorm.Scope) (err error) {

	switch scope.Value.(type) {
	case *Test:

		val := scope.Value.(*Test)

		

		
		if val.UpdateTime.Unix() < 0 {
			val.UpdateTime = time.Now()
		}
		
	}

	return
}



//自动添加更新时间  没有trim
func (this *Test) BeforeSave(scope *gorm.Scope) (err error) {
	val, ok := scope.InstanceGet("gorm:update_attrs")
	if ok {
		switch val.(type) {
		case map[string]interface{}:
			mapVal := val.(map[string]interface{})
			
			if _, ok := mapVal["update_time"]; !ok {
				mapVal["update_time"] = time.Now()
			}
			
		}
	}
	return
}

