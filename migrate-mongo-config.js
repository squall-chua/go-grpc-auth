const config = {
  mongodb: {
    url: process.env.MONGO_URI || "mongodb://localhost:27017/auth_db",
    options: {
      useNewUrlParser: true,
      useUnifiedTopology: true,
    },
  },
  migrationsDir: "migrations",
  changelogCollectionName: "changelog",
  migrationFileExtension: ".js",
  moduleSystem: "commonjs",
};

module.exports = config;
