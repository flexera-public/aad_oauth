build:
	go build -o oauther

run: build
	./oauther --client=$(CLIENT_ID) --secret=$(CLIENT_SECRET) --tenant=$(TENANT_ID) --redirect=https://ad.test.rightscale.com

test:
	curl -H "Authorization: Bearer $(ACCESS_TOKEN)" https://management.azure.com/subscriptions/$(SUBSCRIPTION_ID)/providers?api-version=2014-04-01-preview
