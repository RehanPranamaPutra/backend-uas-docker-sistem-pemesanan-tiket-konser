<?php

namespace App\Http\Controllers;

use App\Models\Concert;
use Illuminate\Http\Request;

class ConcertController extends Controller
{
    public function index()
    {
        return Concert::orderBy('date', 'asc')->get();
    }

    public function show($id)
    {
        return Concert::findOrFail($id);
    }

    public function store(Request $request)
    {
        $data = $request->validate([
            'name' => 'required|string',
            'location' => 'required|string',
            'date' => 'required|date',
            'price' => 'required|integer|min:0',
            'stock' => 'required|integer|min:0',
            'image' => 'nullable|image|mimes:jpeg,png,jpg|max:2048', // Validasi Gambar
        ]);

        // Cek apakah ada file gambar yang diupload
        if ($request->hasFile('image')) {
            // Simpan ke folder 'public/concerts' dan ambil path-nya
            $path = $request->file('image')->store('concerts', 'public');
            $data['image'] = $path;
        }

        return response()->json(Concert::create($data), 201);
    }

    public function updateStock(Request $request, $id)
    {
        // 1. Validasi agar 'reduce_by' harus ada dan berupa angka
        $request->validate([
            'reduce_by' => 'required|integer|min:1'
        ]);

        $concert = Concert::findOrFail($id);

        // 2. LOGIKA PENTING: Kurangi stok yang ada dengan jumlah yang dibeli
        // Jangan gunakan: $concert->stock = $request->stock (karena ini menimpa/mengganti)
        $concert->stock = $concert->stock - $request->reduce_by;

        $concert->save();

        return response()->json([
            'message' => 'Stock updated successfully',
            'remaining_stock' => $concert->stock
        ]);
    }

    public function update(Request $request, $id)
    {
        $concert = Concert::findOrFail($id);

        // Note: Gunakan 'sometimes' agar user tidak wajib upload gambar baru saat edit
        $data = $request->validate([
            'name' => 'required|string',
            'location' => 'required|string',
            'date' => 'required|date',
            'price' => 'required|integer|min:0',
            'stock' => 'required|integer|min:0',
            'image' => 'nullable|image|mimes:jpeg,png,jpg|max:2048',
        ]);

        if ($request->hasFile('image')) {
            // Hapus gambar lama jika ada (agar server tidak penuh sampah)
            if ($concert->image && \Illuminate\Support\Facades\Storage::disk('public')->exists($concert->image)) {
                \Illuminate\Support\Facades\Storage::disk('public')->delete($concert->image);
            }

            // Simpan gambar baru
            $path = $request->file('image')->store('concerts', 'public');
            $data['image'] = $path;
        }

        $concert->update($data);

        return response()->json($concert);
    }

    // --- TAMBAHAN BARU: HAPUS ---
    public function destroy($id)
    {
        $concert = Concert::findOrFail($id);
        $concert->delete();

        return response()->json(['message' => 'Concert deleted successfully']);
    }
}
