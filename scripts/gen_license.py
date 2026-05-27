#!/usr/bin/env python3
"""通过 Ed25519 双层签名方案重新生成 conf/license.dat"""

import base64
import json
import os
from cryptography.hazmat.primitives.asymmetric import ed25519
from cryptography.hazmat.primitives import serialization

# 根私钥（hex 128 字符 = 64 字节 Ed25519 种子）
ROOT_HEX = "8b1786f00ae99d018a0bff45bbab43502da35905a67ac84e6c6ce35945e0abd8ec318b2fa41fdea387889c9ef4ebecfb79bc90ca8af9a4e3f3f704e76f234bee"
# 子私钥
CHILD_HEX = "9bd159d38e0390d9da75d404411fdf295e7b9e988a6fcf95d3ddeb2a2f494e510033bf6e49a1624c2c9ed9db6d4ba87ed3d88b1cfb4641744792180574a239ee"

LICENSE = {
    "customer": "演示版",
    "product_name": "Contful",
    "product_version": "企业版 1.3.0",
    "product_code": "contful-ent-001",
    "is_trial": True,
    "issued_date": "2026-06-01T00:00:00Z",
    "expiry_date": "2027-12-30T00:00:00Z",
}

OUTPUT = os.path.join(os.path.dirname(__file__), "..", "conf", "license.dat")


def main():
    root_seed = bytes.fromhex(ROOT_HEX)
    child_seed = bytes.fromhex(CHILD_HEX)

    root_priv = ed25519.Ed25519PrivateKey.from_private_bytes(root_seed)
    child_priv = ed25519.Ed25519PrivateKey.from_private_bytes(child_seed)

    # 提取子公钥
    child_pub = child_priv.public_key()
    child_pub_raw = child_pub.public_bytes(
        serialization.Encoding.Raw, serialization.PublicFormat.Raw
    )

    # ① 根私钥签子公钥
    root_sig = root_priv.sign(child_pub_raw)

    # ② 序列化 License
    license_json = json.dumps(LICENSE, ensure_ascii=False).encode("utf-8")

    # ③ 子私钥签 License
    child_sig = child_priv.sign(license_json)

    # ④ 打包: 子公钥.根签名.License JSON.子签名
    auth = ".".join([
        base64.b64encode(child_pub_raw).decode(),
        base64.b64encode(root_sig).decode(),
        base64.b64encode(license_json).decode(),
        base64.b64encode(child_sig).decode(),
    ])

    os.makedirs(os.path.dirname(OUTPUT), exist_ok=True)
    with open(OUTPUT, "w") as f:
        f.write(auth)

    print(f"✅ license.dat 已生成: {OUTPUT}")
    print(f"   长度: {len(auth)} 字符")
    print(f"   格式: 4 段 base64 以 . 分隔")
    print()

    # 验证：用根公钥验证子公钥签名
    root_pub_key = root_priv.public_key()
    try:
        root_pub_key.verify(root_sig, child_pub_raw)
        print("✅ 根签名验证: 通过")
    except Exception as e:
        print(f"❌ 根签名验证: 失败 - {e}")

    # 验证：用子公钥验证 License 签名
    try:
        child_pub.verify(child_sig, license_json)
        print("✅ 子签名验证: 通过")
    except Exception as e:
        print(f"❌ 子签名验证: 失败 - {e}")


if __name__ == "__main__":
    main()
