.PHONY: build

build:
	go build -o ./nyaa-dl.out cmd/nyaa-dl/main.go

TAG =
release:
	git checkout develop;
	git push origin develop;
	git checkout main;
	git merge --no-edit develop;
	git push origin main
	git tag ${TAG}
	git push origin tag ${TAG}
	git checkout develop
