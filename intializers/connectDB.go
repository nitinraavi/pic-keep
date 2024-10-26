package intializers

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectToDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(os.Getenv("DBURI")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		panic(err)
		// panic("Failed to Connect to DB: %v", err)
	}

}
