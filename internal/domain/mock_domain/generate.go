package mock_domain

//go:generate mockgen -source=../comment.go -destination=./comment.go
//go:generate mockgen -source=../fetch.go -destination=./fetch.go
//go:generate mockgen -source=../message.go -destination=./message.go
//go:generate mockgen -source=../recommend.go -destination=./recommend.go
//go:generate mockgen -source=../profile.go -destination=./profile.go
//go:generate mockgen -source=../validator.go -destination=./validator.go
