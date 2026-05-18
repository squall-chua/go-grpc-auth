module.exports = {
  async up(db) {
    const now = new Date();

    // Create default namespace
    await db.collection("namespaces").insertOne({
      name: "default",
      config: {
        mfa_required: false,
        allowed_social_providers: [],
        password_policy: {
          min_length: 8,
          require_uppercase: false,
          require_lowercase: false,
          require_number: false,
          require_special: false,
          password_history: 0,
        },
        ip_allowlist: [],
        ip_denylist: [],
        webhook_url: "",
        webhook_secret: "",
      },
    });

    // Create default roles
    await db.collection("roles").insertMany([
      {
        name: "superadmin",
        namespace: "default",
        permissions: [
          "users:read",
          "users:write",
          "roles:read",
          "roles:write",
          "permissions:read",
          "permissions:write",
          "namespaces:read",
          "namespaces:write",
          "clients:read",
          "clients:write",
          "audit:read",
        ],
        created_at: now,
        updated_at: now,
      },
      {
        name: "admin",
        namespace: "default",
        permissions: [
          "users:read",
          "users:write",
          "roles:read",
          "audit:read",
        ],
        created_at: now,
        updated_at: now,
      }
    ]);

    // Create default permissions
    await db.collection("permissions").insertMany([
      { name: "users:read", namespace: "default", description: "Read user data", created_at: now, updated_at: now },
      { name: "users:write", namespace: "default", description: "Create and modify users", created_at: now, updated_at: now },
      { name: "roles:read", namespace: "default", description: "Read roles", created_at: now, updated_at: now },
      { name: "roles:write", namespace: "default", description: "Create and modify roles", created_at: now, updated_at: now },
      { name: "permissions:read", namespace: "default", description: "Read permissions", created_at: now, updated_at: now },
      { name: "permissions:write", namespace: "default", description: "Create and modify permissions", created_at: now, updated_at: now },
      { name: "namespaces:read", namespace: "default", description: "Read namespaces", created_at: now, updated_at: now },
      { name: "namespaces:write", namespace: "default", description: "Create and modify namespaces", created_at: now, updated_at: now },
      { name: "clients:read", namespace: "default", description: "Read OIDC clients", created_at: now, updated_at: now },
      { name: "clients:write", namespace: "default", description: "Create and modify OIDC clients", created_at: now, updated_at: now },
      { name: "audit:read", namespace: "default", description: "Read audit logs", created_at: now, updated_at: now },
    ]);

    // Create superadmin user (password: Admin@123)
    const passwordHash = "$2a$10$xT4JYmwhXm1umlJJ4ihrJeYAodxRf8C7BzxoQSs7lsgnTKcvqufCe";
    await db.collection("users").insertOne({
      email: "admin@localhost",
      username: "superadmin",
      password_hash: passwordHash,
      status: "active",
      roles: ["superadmin"],
      permissions: [],
      password_history: [passwordHash],
      namespace: "default",
      social_identities: [],
      created_at: now,
      updated_at: now,
    });
  },

  async down(db) {
    await db.collection("users").deleteOne({ username: "superadmin", namespace: "default" });
    await db.collection("permissions").deleteMany({ namespace: "default" });
    await db.collection("roles").deleteMany({ namespace: "default" });
    await db.collection("namespaces").deleteOne({ name: "default" });
  },
};
