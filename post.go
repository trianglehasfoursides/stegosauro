package main

type post struct {
	Title   string `gorm:"unique"`
	Desc    string
	Content string
}
