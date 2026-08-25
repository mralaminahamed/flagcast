// flagcast — Mongo bootstrap. Runs once on first container start.
db = db.getSiblingDB('flagcast');

// Flag key is stored as _id, which is unique by definition — no extra index.
db.createCollection('flags');

db.createCollection('audit');
db.audit.createIndex({ flag_key: 1, timestamp: -1 });
// TTL: keep the audit log 180 days.
db.audit.createIndex({ timestamp: 1 }, { expireAfterSeconds: 15552000 });

print('flagcast: mongo initialized (flags, audit + indexes)');
