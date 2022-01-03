BUILD_FLAGS = "-extldflags '-static'"
GC_FLAGS = '-N -l'

.PHONY: build_all
build_all: build_chanel_wait build_chanel_send

.PHONY: build_chanel_wait
build_chanel_wait:
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/chanel_wait -gcflags=${GC_FLAGS} -ldflags ${BUILD_FLAGS} ./cmd/chanel_wait/

.PHONY: build_chanel_send
build_chanel_send:
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/chanel_send_main -gcflags=${GC_FLAGS} -ldflags ${BUILD_FLAGS} ./cmd/chanel_send/