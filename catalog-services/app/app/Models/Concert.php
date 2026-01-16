<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class Concert extends Model
{
        protected $fillable = [
        'name',
        'location',
        'date',
        'price',
        'stock',
        'image',
    ];

}
