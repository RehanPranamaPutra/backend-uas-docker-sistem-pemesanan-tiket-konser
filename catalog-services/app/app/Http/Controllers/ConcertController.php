<?php

namespace App\Http\Controllers;

use Illuminate\Support\Facades\Storage;
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
        $validated = $request->validate([
            'name' => 'required|string',
            'location' => 'required|string',
            'date' => 'required|date',
            'price' => 'required|integer|min:0',
            'stock' => 'required|integer|min:0',
            'image' => 'nullable|image|mimes:jpg,jpeg,png|max:2048'

        ]);

        if ($request->hasFile('image')) {
        $validated['image'] = $request->file('image')
            ->store('concerts', 'public');
        }

        $concert = Concert::create($validated);

        return response()->json([
            'message' => 'Concert created',
            'data' => $concert
         ], 201);
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

        // Validasi: semua optional (partial update data)
        $validated = $request->validate([
            'name' => 'sometimes|required|string',
            'location' => 'sometimes|required|string',
            'date' => 'sometimes|required|date',
            'price' => 'sometimes|required|integer|min:0',
            'stock' => 'sometimes|required|integer|min:0',
            'image' => 'nullable|image|mimes:jpg,jpeg,png|max:2048'
        ]);

        // Kalau ada upload image baru, hapus image lama dulu
        if ($request->hasFile('image')) {
            if (!empty($concert->image) && Storage::disk('public')->exists($concert->image)) {
                Storage::disk('public')->delete($concert->image);
            }

            $validated['image'] = $request->file('image')->store('concerts', 'public');
        }

        if (empty($validated) && !$request->hasFile('image')) {
            return response()->json([
                'message' => 'No fields to update (request body not received)',
                'received' => $request->all(),
            ], 400);
        }


        $concert->update($validated);

        $concert->refresh();

        return response()->json([
            'message' => 'Concert updated successfully',
            'data' => $concert
        ]);
    }

    public function destroy($id)
    {
        $concert = Concert::findOrFail($id);

        // Hapus file image kalau ada
        if (!empty($concert->image) && Storage::disk('public')->exists($concert->image)) {
            Storage::disk('public')->delete($concert->image);
        }

        $concert->delete();

        return response()->json([
            'message' => 'Concert deleted successfully'
        ]);
    }

}
