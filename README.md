Ulaminator
==========

Ulaminator is a small application that will generate a square Ulam spiral
and save it to a PNG file.  By default, the spiral will be greyscale, where
the darkness of each pixel is determined by the number of prime factors of
the number that pixel is associated with.  You can also set it to save in
monochrome, where prime numbers have a black pixel. The program is threaded
so it benefits from a CPU with multiple cores.

This project was last updated Feb 2012, and moved from bitbucket to
github Aug 2013.  I plan on doing a cleanup, and maybe splitting some
functionality into a go get-able package.
