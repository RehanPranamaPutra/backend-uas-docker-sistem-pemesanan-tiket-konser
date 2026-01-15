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
        $request->validate([
            'stock' => 'required|integer|min:0'
        ]);

        $concert = Concert::findOrFail($id);
        $concert->stock = $request->stock;
        $concert->save();

        return $concert;
    }
}
