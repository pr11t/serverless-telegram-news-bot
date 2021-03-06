variable "project" {}
variable "region" {}
variable "prefix" {}
variable "scheduler_cron" {}
variable "telegram_bot_token" {}
variable "telegram_chat_id" {}
variable "rss_feed_url" {}

provider "google" {
  project = var.project
  region  = var.region
}

data "archive_file" "init" {
  type        = "zip"
  source_file = "../rssbot/function.go"
  output_path = "function.zip"
}

resource "google_storage_bucket" "bucket" {
  name     = "${var.prefix}-function-deploy-bucket"
  location = var.region
}

resource "google_storage_bucket_object" "archive" {
  name   = "function.zip"
  bucket = google_storage_bucket.bucket.name
  source = "function.zip"
}

resource "google_service_account" "function_account" {
  account_id = "${var.prefix}-function-runner"
}

resource "google_cloudfunctions_function" "function" {
  name                  = "${var.prefix}-function"
  runtime               = "go111"
  entry_point           = "Run"
  available_memory_mb   = 128
  timeout               = 60
  max_instances         = 1
  ingress_settings      = "ALLOW_ALL"
  service_account_email = google_service_account.function_account.email
  source_archive_bucket = google_storage_bucket.bucket.name
  source_archive_object = google_storage_bucket_object.archive.name
  trigger_http          = true
  environment_variables = {
    TELEGRAM_BOT_TOKEN = var.telegram_bot_token
    TELEGRAM_CHAT_ID   = var.telegram_chat_id
    RSS_FEED_URL       = var.rss_feed_url
  }
}


resource "google_cloudfunctions_function_iam_member" "invoker" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.invoker"
  member = "serviceAccount:${google_service_account.function_account.email}"
}

resource "google_cloud_scheduler_job" "job" {
  name      = "${var.prefix}-scheduler"
  schedule  = var.scheduler_cron
  time_zone = "UTC"
  http_target {
    http_method = "GET"
    uri         = google_cloudfunctions_function.function.https_trigger_url

    oidc_token {
      service_account_email = google_service_account.function_account.email
    }
  }
}

resource "google_service_account" "deployer_account" {
  account_id = "${var.prefix}-function-deployer"
  provisioner "local-exec" {
    # TODO: Generate this with google_service_account_key and output in required format
    command = "gcloud iam service-accounts keys create ${google_service_account.deployer_account.email}.json --iam-account ${google_service_account.deployer_account.email}"
  }
}

resource "google_cloudfunctions_function_iam_member" "deployer" {
  project        = google_cloudfunctions_function.function.project
  region         = google_cloudfunctions_function.function.region
  cloud_function = google_cloudfunctions_function.function.name

  role   = "roles/cloudfunctions.developer"
  member = "serviceAccount:${google_service_account.deployer_account.email}"
}


resource "google_service_account_iam_member" "deployer-account-iam" {
  service_account_id = google_service_account.function_account.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${google_service_account.deployer_account.email}"
}

resource "google_project_iam_custom_role" "cloud_function_deployer" {
  role_id     = "cloudfunctionssourceCodeSet"
  title       = "Cloud function deployer"
  description = "Allow setting source code of cloud function"
  permissions = ["cloudfunctions.functions.sourceCodeSet"]
}

resource "google_project_iam_member" "project" {
  project = var.project
  role    = google_project_iam_custom_role.cloud_function_deployer.name
  member  = "serviceAccount:${google_service_account.deployer_account.email}"
}
