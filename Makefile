.PHONY: build test testacc fmt vet tidy docs install

BINARY := terraform-provider-sendgrid

build:
	go build -o $(BINARY) .

test:
	go test ./...

testacc:
	TF_ACC=1 go test ./internal/provider/... -timeout 30m

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate \
		--provider-name sendgrid

install: build
	mkdir -p $(HOME)/.local/share/terraform/plugins/registry.terraform.io/weesp-ai/sendgrid/0.0.0-dev/$(shell go env GOOS)_$(shell go env GOARCH)
	cp $(BINARY) $(HOME)/.local/share/terraform/plugins/registry.terraform.io/weesp-ai/sendgrid/0.0.0-dev/$(shell go env GOOS)_$(shell go env GOARCH)/
