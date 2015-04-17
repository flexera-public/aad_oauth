build:
	go build -o oauther

run: build
	./oauther --multi --client=$(CLIENT_ID) --secret=$(CLIENT_SECRET) --tenant=$(TENANT_ID) --redirect=https://ad.test.rightscale.com
