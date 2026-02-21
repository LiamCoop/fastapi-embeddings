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

export async function uploadFileToStorage(
  key: string,
  body: Buffer,
  contentType: string,
): Promise<string> {
  const endpoint = getEnv("RAILWAY_BUCKET_ENDPOINT");
  const bucket = getEnv("RAILWAY_BUCKET_NAME");
  const region = process.env.RAILWAY_BUCKET_REGION ?? "auto";
  const accessKeyId = getEnv("RAILWAY_BUCKET_ACCESS_KEY_ID");
  const secretAccessKey = getEnv("RAILWAY_BUCKET_SECRET_ACCESS_KEY");

  const baseUrl = new URL(endpoint);
  const now = new Date();
  const amzDate = toAmzDate(now);
  const dateStamp = amzDate.slice(0, 8);
  const payloadHash = hashSHA256Hex(body);

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
    "PUT",
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
