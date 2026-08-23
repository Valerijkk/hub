variable database_url {
  type    = string
  default = getenv(DATABASE_URL)
}

variable dev_url {
  type    = string
  default = getenv(ATLAS_DEV_URL)
}

env local {
  url = var.database_url
  dev = var.dev_url
  schema {
    src = ent://ent/schema
  }
  migration {
    dir = file://migrations/atlas/mail-calendar-v1
  }
}

env prefeature_baseline {
  dev = var.dev_url
  schema {
    src = ent:///opt/hub/prefeature/backend/ent/schema
  }
}
