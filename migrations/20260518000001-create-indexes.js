module.exports = {
  async up(db) {
    // users
    await db.collection("users").createIndex({ namespace: 1, email: 1 }, { unique: true });
    await db.collection("users").createIndex({ namespace: 1, username: 1 }, { unique: true });

    // tokens
    await db.collection("tokens").createIndex({ token_hash: 1 }, { unique: true });
    await db.collection("tokens").createIndex({ user_id: 1 });
    await db.collection("tokens").createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 });

    // sessions
    await db.collection("sessions").createIndex({ user_id: 1 });
    await db.collection("sessions").createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 });

    // auth_codes
    await db.collection("auth_codes").createIndex({ code: 1 }, { unique: true });
    await db.collection("auth_codes").createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 });

    // mfa_secrets
    await db.collection("mfa_secrets").createIndex({ user_id: 1, method: 1 }, { unique: true });

    // mfa_tokens
    await db.collection("mfa_tokens").createIndex({ token: 1 }, { unique: true });
    await db.collection("mfa_tokens").createIndex({ expires_at: 1 }, { expireAfterSeconds: 0 });

    // oidc_clients
    await db.collection("oidc_clients").createIndex({ client_id: 1 }, { unique: true });
    await db.collection("oidc_clients").createIndex({ namespace: 1 });

    // namespaces
    await db.collection("namespaces").createIndex({ name: 1 }, { unique: true });

    // roles
    await db.collection("roles").createIndex({ namespace: 1, name: 1 }, { unique: true });

    // permissions
    await db.collection("permissions").createIndex({ namespace: 1 });

    // consents
    await db.collection("consents").createIndex({ user_id: 1, client_id: 1 }, { unique: true });

    // audit_logs
    await db.collection("audit_logs").createIndex({ namespace: 1, timestamp: -1 });
    await db.collection("audit_logs").createIndex({ user_id: 1, timestamp: -1 });
    await db.collection("audit_logs").createIndex({ event: 1 });
  },

  async down(db) {
    await db.collection("users").dropIndex("namespace_1_email_1");
    await db.collection("users").dropIndex("namespace_1_username_1");

    await db.collection("tokens").dropIndex("token_hash_1");
    await db.collection("tokens").dropIndex("user_id_1");
    await db.collection("tokens").dropIndex("expires_at_1");

    await db.collection("sessions").dropIndex("user_id_1");
    await db.collection("sessions").dropIndex("expires_at_1");

    await db.collection("auth_codes").dropIndex("code_1");
    await db.collection("auth_codes").dropIndex("expires_at_1");

    await db.collection("mfa_secrets").dropIndex("user_id_1_method_1");

    await db.collection("mfa_tokens").dropIndex("token_1");
    await db.collection("mfa_tokens").dropIndex("expires_at_1");

    await db.collection("oidc_clients").dropIndex("client_id_1");
    await db.collection("oidc_clients").dropIndex("namespace_1");

    await db.collection("namespaces").dropIndex("name_1");

    await db.collection("roles").dropIndex("namespace_1_name_1");

    await db.collection("permissions").dropIndex("namespace_1");

    await db.collection("consents").dropIndex("user_id_1_client_id_1");

    await db.collection("audit_logs").dropIndex("namespace_1_timestamp_-1");
    await db.collection("audit_logs").dropIndex("user_id_1_timestamp_-1");
    await db.collection("audit_logs").dropIndex("event_1");
  },
};
