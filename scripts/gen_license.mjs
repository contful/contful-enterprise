#!/usr/bin/env node
/** 通过 Ed25519 双层签名方案重新生成 conf/license.dat */

import { createPrivateKey, createPublicKey, sign, verify } from 'crypto'
import { readFileSync, writeFileSync, mkdirSync } from 'fs'
import { dirname, join } from 'path'
import { fileURLToPath } from 'url'

const __dirname = dirname(fileURLToPath(import.meta.url))

// 根私钥 seed (64 bytes hex = Ed25519 32-byte seed)
const ROOT_SEED = Buffer.from(
  '8b1786f00ae99d018a0bff45bbab43502da35905a67ac84e6c6ce35945e0abd8ec318b2fa41fdea387889c9ef4ebecfb79bc90ca8af9a4e3f3f704e76f234bee',
  'hex'
)
// 子私钥 seed
const CHILD_SEED = Buffer.from(
  '9bd159d38e0390d9da75d404411fdf295e7b9e988a6fcf95d3ddeb2a2f494e510033bf6e49a1624c2c9ed9db6d4ba87ed3d88b1cfb4641744792180574a239ee',
  'hex'
)

const LICENSE = {
  customer: '演示版',
  product_name: 'Contful',
  product_version: '企业版 1.3.0',
  product_code: 'contful-ent-001',
  is_trial: true,
  issued_date: '2026-06-01T00:00:00Z',
  expiry_date: '2027-12-30T00:00:00Z',
}

// 从 seed 创建 Ed25519 KeyObject (使用 DER SPKI/PKCS8 格式)
function ed25519KeyFromSeed(seed) {
  // Node.js crypto 不直接支持 seed 创建 Ed25519 key
  // 使用 PKCS8 DER 格式包装 seed
  // Ed25519 PKCS8: 302e020100300506032b657004220420 + seed(32)
  const pkcs8Prefix = Buffer.from('302e020100300506032b657004220420', 'hex')
  const der = Buffer.concat([pkcs8Prefix, seed])
  return createPrivateKey({ key: der, format: 'der', type: 'pkcs8' })
}

const rootPriv = ed25519KeyFromSeed(ROOT_SEED.subarray(0, 32)) // Ed25519 seed is first 32 bytes
const childPriv = ed25519KeyFromSeed(CHILD_SEED.subarray(0, 32))

// 提取子公钥 (SPKI DER -> raw 32 bytes)
const childPubKey = createPublicKey(childPriv)
const childPubDer = childPubKey.export({ format: 'der', type: 'spki' })
// Ed25519 SPKI: 302a300506032b6570032100 + rawPub(32)
const childPubRaw = childPubDer.subarray(-32)

// ① 根私钥签子公钥
const rootSig = sign(undefined, childPubRaw, rootPriv)

// ② 序列化 License
const licenseJSON = Buffer.from(JSON.stringify(LICENSE), 'utf-8')

// ③ 子私钥签 License
const childSig = sign(undefined, licenseJSON, childPriv)

// ④ 打包: 子公钥.根签名.License JSON.子签名
const auth = [
  childPubRaw.toString('base64'),
  rootSig.toString('base64'),
  licenseJSON.toString('base64'),
  childSig.toString('base64'),
].join('.')

const output = join(__dirname, '..', 'conf', 'license.dat')
mkdirSync(dirname(output), { recursive: true })
writeFileSync(output, auth)

console.log(`✅ license.dat 已生成: ${output}`)
console.log(`   长度: ${auth.length} 字符`)
console.log(`   格式: 4 段 base64 以 . 分隔\n`)

// 验证
const rootPubKey = createPublicKey(rootPriv)
const verifyRoot = verify(undefined, childPubRaw, rootPubKey, rootSig)
console.log(`✅ 根签名验证: ${verifyRoot ? '通过' : '失败'}`)

const verifyChild = verify(undefined, licenseJSON, childPubKey, childSig)
console.log(`✅ 子签名验证: ${verifyChild ? '通过' : '失败'}`)
