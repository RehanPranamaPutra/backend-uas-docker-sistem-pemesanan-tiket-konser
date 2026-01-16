<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::table('concerts', function (Blueprint $table) {
            // Tambahkan kolom image setelah kolom location
            // nullable() artinya boleh kosong (biar data lama gak error)
            $table->string('image')->nullable()->after('location');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('concerts', function (Blueprint $table) {
            $table->dropColumn('image');
        });
    }
};
