include $(GOROOT)/src/Make.inc

TARG=mpack

GOFILES=constants.go pack_writer.go pack_reader.go mpack.go rpc.go array.go map.go

include $(GOROOT)/src/Make.pkg

format:
	gofmt -w=true -l=true -spaces=true *.go

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O
