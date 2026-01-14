const express = require("express");
const bcrypt = require("bcrypt");
const jwt = require("jsonwebtoken");
const rateLimit = require("express-rate-limit");
const { z } = require("zod");

const User = require("../models/User");
const HttpError = require("../utils/httpError");
const { authRequired } = require("../middlewares/auth.middleware");

const router = express.Router();

const loginLimiter = rateLimit({
  windowMs: 60 * 1000,
  max: 10,
  standardHeaders: true,
  legacyHeaders: false,
});

const registerSchema = z.object({
  username: z.string().min(3),
  email: z.string().email(),
  password: z.string().min(6),
  role: z.enum(["user", "admin"]).optional(),
});

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(6),
});

router.get("/health", (req, res) => res.json({ status: "ok", service: "user-service" }));

router.post("/auth/register", async (req, res, next) => {
  try {
    const data = registerSchema.parse(req.body);

    const exists = await User.findOne({ email: data.email });
    if (exists) throw new HttpError(409, "Email already registered");

    const passwordHash = await bcrypt.hash(data.password, 10);

    const user = await User.create({
      username: data.username,
      email: data.email,
      passwordHash,
      role: data.role || "user",
    });

    res.status(201).json({
      id: user._id,
      username: user.username,
      email: user.email,
      role: user.role,
    });
  } catch (err) {
    next(err);
  }
});

router.post("/auth/login", loginLimiter, async (req, res, next) => {
  try {
    const data = loginSchema.parse(req.body);

    const user = await User.findOne({ email: data.email });
    if (!user) throw new HttpError(401, "Invalid email or password");

    const ok = await bcrypt.compare(data.password, user.passwordHash);
    if (!ok) throw new HttpError(401, "Invalid email or password");

    const token = jwt.sign(
      { sub: user._id.toString(), role: user.role },
      process.env.JWT_SECRET,
      { expiresIn: process.env.JWT_EXPIRES_IN || "1d" }
    );

    res.json({
      token,
      user: { id: user._id, username: user.username, email: user.email, role: user.role },
    });
  } catch (err) {
    next(err);
  }
});

router.get("/me", authRequired, async (req, res, next) => {
  try {
    const user = await User.findById(req.user.sub).select("-passwordHash");
    if (!user) throw new HttpError(404, "User not found");
    res.json(user);
  } catch (err) {
    next(err);
  }
});

module.exports = router;
