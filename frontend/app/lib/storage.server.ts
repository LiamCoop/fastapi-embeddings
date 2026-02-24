import "server-only";

import { createHash, createHmac } from "node:crypto";

function getEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required environment variable: ${name}`);
  }
  return value;
}

function hashSHA256Hex(value: string | Buffer): string {
  return createHash("sha256").update(value).digest("hex");
}

function hmacSHA256(key: Buffer | string, value: string): Buffer {
  return createHmac("sha256", key).update(value).digest();
}

function toAmzDate(date: Date): string {
  return date.toISOString().replace(/[:-]|\.\d{3}/g, "");
}

function encodePath(path: string): string {
  return path
    .split("/")
    .map((segment) => encodeURIComponent(segment))
    .join("/");
}

type SignedStorageRequest = {
  requestUrl: URL;
  authorization: string;
  amzDate: string;
  payloadHash: string;
};

function buildSignedStorageRequest(
  method: "PUT" | "DELETE",
  key: string,
  payloadHash: string,
  now: Date,
): SignedStorageRequest {
  const endpoint = getEnv("RAILWAY_BUCKET_ENDPOINT");
  const bucket = getEnv("RAILWAY_BUCKET_NAME");
  const region = process.env.RAILWAY_BUCKET_REGION ?? "auto";
  const accessKeyId = getEnv("RAILWAY_BUCKET_ACCESS_KEY_ID");
  const secretAccessKey = getEnv("RAILWAY_BUCKET_SECRET_ACCESS_KEY");

  const baseUrl = new URL(endpoint);
  const amzDate = toAmzDate(now);
  const dateStamp = amzDate.slice(0, 8);

  const keyPath = key.replace(/^\/+/, "");
  const endpointPrefix = baseUrl.pathname.replace(/^\/+|\/+$/g, "");
  const objectPath = endpointPrefix
    ? `${endpointPrefix}/${bucket}/${keyPath}`
    : `${bucket}/${keyPath}`;
  const canonicalUri = `/${encodePath(objectPath)}`;
  const host = baseUrl.host;

  const canonicalHeaders = [
    `host:${host}`,
    `x-amz-content-sha256:${payloadHash}`,
    `x-amz-date:${amzDate}`,
    "",
  ].join("\n");

  const signedHeaders = "host;x-amz-content-sha256;x-amz-date";
  const canonicalRequest = [
    method,
    canonicalUri,
    "",
    canonicalHeaders,
    signedHeaders,
    payloadHash,
  ].join("\n");

  const credentialScope = `${dateStamp}/${region}/s3/aws4_request`;
  const stringToSign = [
    "AWS4-HMAC-SHA256",
    amzDate,
    credentialScope,
    hashSHA256Hex(canonicalRequest),
  ].join("\n");

  const signingKey = hmacSHA256(
    hmacSHA256(hmacSHA256(hmacSHA256(`AWS4${secretAccessKey}`, dateStamp), region), "s3"),
    "aws4_request",
  );
  const signature = createHmac("sha256", signingKey).update(stringToSign).digest("hex");
  const authorization = [
    `AWS4-HMAC-SHA256 Credential=${accessKeyId}/${credentialScope}`,
    `SignedHeaders=${signedHeaders}`,
    `Signature=${signature}`,
  ].join(", ");

  const requestUrl = new URL(baseUrl.toString());
  requestUrl.pathname = canonicalUri;

  return {
    requestUrl,
    authorization,
    amzDate,
    payloadHash,
  };
}

export async function uploadFileToStorage(
  key: string,
  body: Buffer,
  contentType: string,
): Promise<string> {
  const payloadHash = hashSHA256Hex(body);
  const { requestUrl, authorization, amzDate } = buildSignedStorageRequest("PUT", key, payloadHash, new Date());
  const bucket = getEnv("RAILWAY_BUCKET_NAME");
  const keyPath = key.replace(/^\/+/, "");

  const response = await fetch(requestUrl, {
    method: "PUT",
    headers: {
      Authorization: authorization,
      "Content-Type": contentType,
      "x-amz-content-sha256": payloadHash,
      "x-amz-date": amzDate,
    },
    body: new Uint8Array(body),
  });

  if (!response.ok) {
    const bodyText = await response.text();
    throw new Error(`Bucket upload failed (${response.status}): ${bodyText.slice(0, 200)}`);
  }

  return `s3://${bucket}/${keyPath}`;
}

function parseS3Uri(uri: string): { bucket: string; key: string } {
  if (!uri.startsWith("s3://")) {
    throw new Error(`Unsupported storage URI: ${uri}`);
  }

  const withoutScheme = uri.slice("s3://".length);
  const slashIndex = withoutScheme.indexOf("/");
  if (slashIndex === -1) {
    throw new Error(`Invalid storage URI (missing object key): ${uri}`);
  }

  const bucket = withoutScheme.slice(0, slashIndex);
  const key = withoutScheme.slice(slashIndex + 1);
  return { bucket, key };
}

export async function deleteFileFromStorageUri(uri: string): Promise<void> {
  const { bucket, key } = parseS3Uri(uri);
  const configuredBucket = getEnv("RAILWAY_BUCKET_NAME");

  if (bucket !== configuredBucket) {
    throw new Error(
      `Storage bucket mismatch for delete: uri bucket ${bucket} does not match configured bucket ${configuredBucket}`,
    );
  }

  const emptyPayloadHash = hashSHA256Hex("");
  const { requestUrl, authorization, amzDate } = buildSignedStorageRequest(
    "DELETE",
    key,
    emptyPayloadHash,
    new Date(),
  );

  const response = await fetch(requestUrl, {
    method: "DELETE",
    headers: {
      Authorization: authorization,
      "x-amz-content-sha256": emptyPayloadHash,
      "x-amz-date": amzDate,
    },
  });

  if (response.status === 404 || response.status === 204) {
    return;
  }

  if (!response.ok) {
    const bodyText = await response.text();
    throw new Error(`Bucket delete failed (${response.status}): ${bodyText.slice(0, 200)}`);
  }
}
