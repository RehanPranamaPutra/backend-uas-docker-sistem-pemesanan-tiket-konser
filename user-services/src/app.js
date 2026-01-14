const express = require("express");
const cors = require("cors");
const helmet = require("helmet");
const { ZodError } = require("zod");
const HttpError = require("./utils/httpError");
const authRoutes = require("./routes/auth.routes");

const app = express();

app.use(helmet());
app.use(cors());
app.use(express.json());

app.use("/", authRoutes);

// 404
app.use((req, res, next) => next(new HttpError(404, "Route not found")));

// Error handler
app.use((err, req, res, next) => {
  if (err instanceof ZodError) {
    return res.status(400).json({ message: "Validation error", issues: err.issues });
  }
  const status = err.statusCode || 500;
  res.status(status).json({ message: err.message || "Internal server error" });
});

module.exports = app;
