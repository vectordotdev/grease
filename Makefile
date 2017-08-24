build_dir := $(CURDIR)/build
dist_dir := $(CURDIR)/dist

exec := grease
github_repo := timberio/grease
version = $(shell cat VERSION)

.DEFAULT_GOAL := dist

.PHONY: clean
clean: clean-build clean-dist

.PHONY: clean-build
clean-build:
	@echo "Removing build files"
	rm -rf $(build_dir)

.PHONY: clean-dist
clean-dist:
	@echo "Removing distribution files"
	rm -rf $(dist_dir)

.PHONY: build
build: clean-build
	@echo "Creating build directory"
	mkdir -p $(build_dir)
	@echo "Building targets"
	@gox -ldflags "-X main.version=$(version)" \
		-osarch="darwin/amd64" \
		-osarch="freebsd/amd64" \
		-osarch="linux/amd64" \
		-osarch="netbsd/amd64" \
		-osarch="openbsd/amd64" \
		-output "$(build_dir)/$(exec)-$(version)-{{.OS}}-{{.Arch}}/$(exec)/bin/$(exec)"
	@for f in $$(ls $(build_dir)); do \
		support_source="$(CURDIR)/support"; \
		support_dest="$(build_dir)/$$f/$(exec)"; \
		echo "Copying $$support_source into $$support_dest"; \
		cp -r $$support_source $$support_dest; \
		readme_source="$(CURDIR)/README.md"; \
		readme_dest="$(build_dir)/$$f/$(exec)/"; \
		echo "Copying $$readme_source into $$readme_dest"; \
		cp $$readme_source $$readme_dest; \
		changelog_source="$(CURDIR)/CHANGELOG.md"; \
		changelog_dest="$(build_dir)/$$f/$(exec)/"; \
		echo "Copying $$changelog_source into $$changelog_dest"; \
		cp $$changelog_source $$changelog_dest; \
		license_source="$(CURDIR)/LICENSE"; \
		license_dest="$(build_dir)/$$f/$(exec)/"; \
		echo "Copying $$license_source into $$license_dest"; \
		cp $$license_source $$license_dest; \
	done

.PHONY: dist
dist: clean-dist build
	@echo "Creating distribution directory"
	mkdir -p $(dist_dir)
	@echo "Creating distribution archives"
	$(eval FILES := $(shell ls $(build_dir)))
	@for f in $(FILES); do \
		echo "Creating distribution archive for $$f"; \
		(cd $(build_dir)/$$f && tar -czf $(dist_dir)/$$f.tar.gz *); \
	done

.PHONY: release
release: dist
	@tag=v$(version); \
	commit=$(git rev-list -n 1 $$tag); \
	name=$$(git show -s $$tag --pretty=tformat:%N | sed -e '4q;d'); \
	changelog=$$(git show -s $$tag --pretty=tformat:%N | sed -e '1,5d'); \
	grease create-release --name "$$name" --notes "$$changelog" --assets "dist/*" $(github_repo) "$$tag" "$$commit"

.PHONY: get-tools
get-tools:
	go get github.com/golang/dep/cmd/dep
	go get github.com/mitchellh/gox
	go get github.com/jstemmer/go-junit-report

.PHONY: test
test:
	@go test -v
