#!/usr/local/bin/gnuplot

set term pdf size 3.236,2 font "Serif, 16"
set output "duration.pdf"
set ylabel "Duration (s)"
set xlabel "Tree degree"
set notitle
set yrange [0:120]
set xrange [5:20000]
set logscale x

f(x) = a * x + (b / log(x)) * (c + d * x)
fit f(x) "10m" using 1:($2/1000) via a, b, c, d

plot "10m" using 1:($2/1000.0):($3/1000.0) notitle with yerrorbars lw 2, \
     f(x) notitle with lines lw 2
