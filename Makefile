darwin:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -o net-report_darwin64 net-report.go

linux:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o net-report_linux64 net-report.go

windows:
	CGO_ENABLED=0 GOARCH=amd64 GOOS=windows go build -o net-report.exe net-report.go

service-account:
	gcloud iam service-accounts create net-test
	gcloud projects add-iam-policy-binding "${GCP_PROJECT_ID}"  \
		--member "serviceAccount:net-test@${GCP_PROJECT_ID}.iam.gserviceaccount.com"  \
		--role "roles/pubsub.publisher"

key:
	gcloud iam service-accounts keys create net-test.json  \
		--iam-account "net-test@${GCP_PROJECT_ID}.iam.gserviceaccount.com"

topic:
	gcloud pubsub topics create "${REPORT_TOPIC}"

pipeline:
	gsutil mb "gs://${DATAFLOW_STORAGE}"

table:
	bq --location=europe-west3 mk \
		--dataset \
		--description "Network ICMP results" \
		"${GCP_PROJECT_ID}:${BQ_DATASET}"

	bq --location=europe-west3 mk \
		-t \
		--description "Network ICMP results" \
		"${BQ_DATASET}.${BQ_TABLE}" \
		rtt_micros:INTEGER,target_ip_addr:STRING,target_host_name:STRING,source_ip_addr:STRING,source_host_name:STRING,packet_size_bytes:INTEGER,packet_ttl:INTEGER,timestamp:DATETIME,unix_time:INTEGER

.PHONY: darwin linux service-account pipeline topic table key
