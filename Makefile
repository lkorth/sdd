# Build the binaries that ship in the plugin. The flags make each build reproducible, so a
# rebuild with no code change leaves git clean.
#
# The binaries go to the skill only after each step passes: the tests, each build, and a run of
# the launcher with the binary for this machine. A failure leaves skills/sdd/bin as it was.
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64
BIN := skills/sdd/bin
STAGE := bin/stage

.PHONY: test binaries
test:
	go vet ./...
	go test -count=1 ./...

binaries: test
	rm -rf $(STAGE)
	mkdir -p $(STAGE)
	@for p in $(PLATFORMS); do \
		echo "build $(STAGE)/sdd-$$p"; \
		CGO_ENABLED=0 GOOS=$${p%-*} GOARCH=$${p#*-} go build -trimpath -buildvcs=false \
			-ldflags='-s -w -buildid=' -o $(STAGE)/sdd-$$p ./cmd/sdd || exit 1; \
	done
	cp $(BIN)/sdd $(STAGE)/sdd
	@echo "run $(STAGE)/sdd features --root testdata/project"
	@$(STAGE)/sdd features --root testdata/project | grep -q '^LNK  ' || \
		{ echo "the launcher or the binary for this machine does not run"; exit 1; }
	@for p in $(PLATFORMS); do cp $(STAGE)/sdd-$$p $(BIN)/sdd-$$p; done
	rm -rf $(STAGE)
	@echo "copied the binaries to $(BIN)"
