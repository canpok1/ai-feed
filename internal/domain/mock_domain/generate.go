package mock_domain

//go:generate mockgen -source=../fetch.go -destination=./fetch.go
//go:generate mockgen -source=../view.go -destination=./view.go
//go:generate mockgen -source=../recommend.go -destination=./recommend.go
