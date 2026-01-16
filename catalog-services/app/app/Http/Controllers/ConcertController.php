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
        ]);

        return response()->json(
            Concert::create($data),
            201
        );
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
}
