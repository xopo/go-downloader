build:
	go build -o downloader main.go

linux:
	GOOS=linux GOARCH=amd64 go build -o downloader main.go
	
run: build
	./server

# sass: 
# 	cd assets && sass -w scss:css 
#
# watchFile:
# 	reflex -s -r '\.(go|html)$$' make run
#
# watchStyle:
# 	reflex -s -r '\.(scss)$$' make sass
#
# watch:
# 	make -j4 watchFile watchStyle

