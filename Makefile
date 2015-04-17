all:
	go build -o oauther

run: all
	./oauther --client=$(CLIENT_ID) --secret=$(CLIENT_SECRET) --tenant=$(TENANT_ID) --redirect=https://ad.test.rightscale.com
