variable dev_url {
  type    = string
  default = getenv("ATLAS_DIFF_DEV_URL")
}

env local {
  dev = var.dev_url
  migration {
    dir = "file://migrations/atlas/mail-calendar-v1"
  }
}
