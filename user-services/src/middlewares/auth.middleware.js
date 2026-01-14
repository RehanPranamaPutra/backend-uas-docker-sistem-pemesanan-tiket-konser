const jwt = require("jsonwebtoken");
const HttpError = require("../utils/httpError");

function authRequired(req, res, next) {
  const header = req.headers.authorization || "";
  const token = header.startsWith("Bearer ") ? header.slice(7) : null;
  if (!token) return next(new HttpError(401, "Missing Bearer token"));

  try {
    const payload = jwt.verify(token, process.env.JWT_SECRET);
    req.user = payload; // { sub, role }
    next();
  } catch {
    next(new HttpError(401, "Invalid token"));
  }
}

module.exports = { authRequired };
