<?php

use Illuminate\Support\Facades\Route;
use App\Http\Controllers\ConcertController;

Route::get('/health', function () {
    return response()->json([
        'status' => 'ok',
        'service' => 'catalog-service',
    ]);
});

Route::get('/concerts', [ConcertController::class, 'index']);
Route::get('/concerts/{id}', [ConcertController::class, 'show']);
Route::post('/concerts', [ConcertController::class, 'store']);
Route::patch('/concerts/{id}/stock', [ConcertController::class, 'updateStock']);
