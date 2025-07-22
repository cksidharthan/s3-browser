export interface ConnectionCredentials {
  endpoint: string;
  access_key: string;
  secret_key: string;
  region: string;
  use_ssl: boolean;
}

export interface ConnectionTestResponse {
  success: boolean;
  message: string;
}

export interface S3Bucket {
  name: string;
  creation_date: string;
}

export interface S3Object {
  key: string;
  size: number;
  etag: string;
  storage_class: string;
}
