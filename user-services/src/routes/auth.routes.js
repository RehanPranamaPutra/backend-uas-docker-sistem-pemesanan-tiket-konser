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

// --- TAMBAHAN BARU: GET ALL USERS (Khusus Admin) ---
router.get("/users", authRequired, async (req, res, next) => {
  try {
    // 1. Cek apakah yang request adalah Admin
    if (req.user.role !== 'admin') {
      throw new HttpError(403, "Access denied. Admins only.");
    }

    // 2. Ambil semua user dari database (kecuali password hash-nya)
    const users = await User.find().select("-passwordHash");

    res.json(users);
  } catch (err) {
    next(err);
  }
});

// --- 1. UPDATE USER (Hanya bisa update diri sendiri) ---
router.put("/users/:id", authRequired, async (req, res, next) => {
  try {
    // Cek apakah ID yang mau diedit == ID orang yang login
    // req.user.sub adalah ID dari token JWT
    if (req.user.sub !== req.params.id) {
      throw new HttpError(403, "Anda hanya boleh mengedit akun sendiri!");
    }

    const { username, email } = req.body;
    
    // Update data
    const updatedUser = await User.findByIdAndUpdate(
      req.params.id,
      { username, email },
      { new: true } // Agar yang dikembalikan adalah data terbaru
    ).select("-passwordHash");

    res.json(updatedUser);
  } catch (err) {
    next(err);
  }
});

// --- 2. DELETE USER (Hanya Admin yang bisa) ---
router.delete("/users/:id", authRequired, async (req, res, next) => {
  try {
    // Cek apakah yang request adalah Admin
    if (req.user.role !== 'admin') {
      throw new HttpError(403, "Hanya Admin yang boleh menghapus user.");
    }

    // Cek user yang mau dihapus
    const targetUser = await User.findById(req.params.id);
    if (!targetUser) throw new HttpError(404, "User tidak ditemukan");

    // PENTING: Admin tidak boleh menghapus sesama Admin (untuk keamanan)
    if (targetUser.role === 'admin') {
      throw new HttpError(403, "Sesama Admin tidak boleh saling menghapus!");
    }

    await User.findByIdAndDelete(req.params.id);
    res.json({ message: "User berhasil dihapus" });
  } catch (err) {
    next(err);
  }
});

module.exports = router;
