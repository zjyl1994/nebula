package vars

import "gorm.io/gorm"

var (
	LISTEN_ADDR string
	DEBUG_MODE  bool

	DB *gorm.DB
)
