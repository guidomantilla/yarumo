package main

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	cdatasource "github.com/guidomantilla/yarumo/datasource"
	cgorm "github.com/guidomantilla/yarumo/datasource/gorm"
)

type product struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func main() {
	ctx := context.Background()

	dsCtx := cdatasource.NewContext(":memory:", "u", "p", "host", "products")

	errCh := make(chan error, 1)

	conn, closeFn, err := cgorm.BuildDB(ctx, dsCtx, cgorm.SqliteOpener(), errCh)
	if err != nil {
		fmt.Println("BuildDB:", err)
		return
	}
	defer closeFn(ctx, 5*time.Second)

	raw, err := conn.Connect(ctx)
	if err != nil {
		fmt.Println("Connect:", err)
		return
	}

	gdb := raw.(*gorm.DB)

	err = gdb.AutoMigrate(&product{})
	if err != nil {
		fmt.Println("AutoMigrate:", err)
		return
	}

	handler := cgorm.NewTransactionHandler(conn)

	err = handler.HandleTransaction(ctx, func(txCtx context.Context) error {
		tx, _ := cgorm.TxFromContext(txCtx)

		return tx.Create(&product{Name: "widget"}).Error
	})

	if err != nil {
		fmt.Println("transaction failed:", err)
		return
	}

	var rows []product
	_ = gdb.Find(&rows).Error

	for _, r := range rows {
		fmt.Printf("row: id=%d name=%s\n", r.ID, r.Name)
	}
}
