VERSION?="dev"

docker: buildprep
	docker build --build-arg=VERSION=$(VERSION) .

buildprep:
	git fetch --tags -f
	mkdir -p dest
	$(eval VERSION=$(shell git describe --tags | cut -d "v" -f 2 | cut -d "-" -f 1))