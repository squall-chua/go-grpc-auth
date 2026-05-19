module.exports = {
  async up(db) {
    await db.collection("permissions").dropIndex("namespace_1");
    await db.collection("permissions").createIndex({ namespace: 1, name: 1 }, { unique: true });
  },

  async down(db) {
    await db.collection("permissions").dropIndex("namespace_1_name_1");
    await db.collection("permissions").createIndex({ namespace: 1 });
  },
};
