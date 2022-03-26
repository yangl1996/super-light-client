#!/usr/local/bin/gnuplot

set term pdf size 3.236,2 font "Serif, 16"
set output "trend.pdf"
set ylabel "Duration (s)"
set xlabel "Ledger size"
set notitle
set yrange [0:30]
set xrange [500:200000000]
set logscale x

f(x) = a + b * log(x)
fit f(x) "trend" using 1:($2/1000) via a, b

plot "trend" using 1:($2/1000.0):($3/1000.0) notitle with yerrorbars lw 2, \
     f(x) notitle with lines lw 2
