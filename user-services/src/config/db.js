const mongoose = require("mongoose");

async function connectDB() {
  const uri = process.env.MONGO_URI;
  if (!uri) throw new Error("MONGO_URI is missing");

  mongoose.set("strictQuery", true);
  await mongoose.connect(uri);
  console.log("✅ Connected to MongoDB");
}

module.exports = connectDB;
