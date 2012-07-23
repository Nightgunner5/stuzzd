package util

// Ported from https://github.com/Bukkit/Bukkit/blob/master/src/main/java/org/bukkit/util/noise/SimplexNoiseGenerator.java

import (
	"math"
)

var (
	sqrt3 = math.Sqrt(3)
	sqrt5 = math.Sqrt(5)
	f2    = 0.5 * (sqrt3 - 1.)
	g2    = (3. - sqrt3) / 6.
	g22   = g2*2. - 1.
	f3    = 1. / 3.
	g3    = 1. / 6.
	f4    = (sqrt5 - 1.) / 4.
	g4    = (5. - sqrt5) / 20.
	g42   = g4 * 2.
	g43   = g4 * 3.
	g44   = g4*4. - 1.
	grad3 = [][]int{
		{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
		{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
		{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
	}
	grad4 = [][]int{
		{0, 1, 1, 1}, {0, 1, 1, -1}, {0, 1, -1, 1}, {0, 1, -1, -1},
		{0, -1, 1, 1}, {0, -1, 1, -1}, {0, -1, -1, 1}, {0, -1, -1, -1},
		{1, 0, 1, 1}, {1, 0, 1, -1}, {1, 0, -1, 1}, {1, 0, -1, -1},
		{-1, 0, 1, 1}, {-1, 0, 1, -1}, {-1, 0, -1, 1}, {-1, 0, -1, -1},
		{1, 1, 0, 1}, {1, 1, 0, -1}, {1, -1, 0, 1}, {1, -1, 0, -1},
		{-1, 1, 0, 1}, {-1, 1, 0, -1}, {-1, -1, 0, 1}, {-1, -1, 0, -1},
		{1, 1, 1, 0}, {1, 1, -1, 0}, {1, -1, 1, 0}, {1, -1, -1, 0},
		{-1, 1, 1, 0}, {-1, 1, -1, 0}, {-1, -1, 1, 0}, {-1, -1, -1, 0},
	}
	perm = []int{
		151, 160, 137, 91, 90, 15, 131, 13, 201,
		95, 96, 53, 194, 233, 7, 225, 140, 36, 103, 30, 69, 142, 8, 99, 37,
		240, 21, 10, 23, 190, 6, 148, 247, 120, 234, 75, 0, 26, 197, 62,
		94, 252, 219, 203, 117, 35, 11, 32, 57, 177, 33, 88, 237, 149, 56,
		87, 174, 20, 125, 136, 171, 168, 68, 175, 74, 165, 71, 134, 139,
		48, 27, 166, 77, 146, 158, 231, 83, 111, 229, 122, 60, 211, 133,
		230, 220, 105, 92, 41, 55, 46, 245, 40, 244, 102, 143, 54, 65, 25,
		63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169, 200,
		196, 135, 130, 116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3,
		64, 52, 217, 226, 250, 124, 123, 5, 202, 38, 147, 118, 126, 255,
		82, 85, 212, 207, 206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42,
		223, 183, 170, 213, 119, 248, 152, 2, 44, 154, 163, 70, 221, 153,
		101, 155, 167, 43, 172, 9, 129, 22, 39, 253, 19, 98, 108, 110, 79,
		113, 224, 232, 178, 185, 112, 104, 218, 246, 97, 228, 251, 34, 242,
		193, 238, 210, 144, 12, 191, 179, 162, 241, 81, 51, 145, 235, 249,
		14, 239, 107, 49, 192, 214, 31, 181, 199, 106, 157, 184, 84, 204,
		176, 115, 121, 50, 45, 127, 4, 150, 254, 138, 236, 205, 93, 222,
		114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180,
	}
	simplex = [][]int{
		{0, 1, 2, 3}, {0, 1, 3, 2}, {0, 0, 0, 0}, {0, 2, 3, 1}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {1, 2, 3, 0},
		{0, 2, 1, 3}, {0, 0, 0, 0}, {0, 3, 1, 2}, {0, 3, 2, 1}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {1, 3, 2, 0},
		{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0},
		{1, 2, 0, 3}, {0, 0, 0, 0}, {1, 3, 0, 2}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {2, 3, 0, 1}, {2, 3, 1, 0},
		{1, 0, 2, 3}, {1, 0, 3, 2}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {2, 0, 3, 1}, {0, 0, 0, 0}, {2, 1, 3, 0},
		{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0},
		{2, 0, 1, 3}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {3, 0, 1, 2}, {3, 0, 2, 1}, {0, 0, 0, 0}, {3, 1, 2, 0},
		{2, 1, 0, 3}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {3, 1, 0, 2}, {0, 0, 0, 0}, {3, 2, 0, 1}, {3, 2, 1, 0},
	}
)

func init() {
	perm = append(perm, perm...)
}

func dot(g []int, h ...float64) float64 {
	var out float64
	for i, j := range h {
		out += float64(g[i]) * j
	}
	return out
}

func Noise3(xin, yin, zin float64) float64 {
	var n0, n1, n2, n3 float64 // Noise contributions from the four corners

	// Skew the input space to determine which simplex cell we're in
	s := (xin + yin + zin) * f3 // Very nice and simple skew factor for 3D
	i := math.Floor(xin + s)
	j := math.Floor(yin + s)
	k := math.Floor(zin + s)

	t := (i + j + k) * g3
	X0 := i - t // Unskew the cell origin back to (x,y,z) space
	Y0 := j - t
	Z0 := k - t
	x0 := xin - X0 // The x,y,z distances from the cell origin
	y0 := yin - Y0
	z0 := zin - Z0

	// For the 3D case, the simplex shape is a slightly irregular tetrahedron.

	// Determine which simplex we are in.
	var i1, j1, k1 int // Offsets for second corner of simplex in (i,j,k) coords
	var i2, j2, k2 int // Offsets for third corner of simplex in (i,j,k) coords
	if x0 >= y0 {
		if y0 >= z0 {
			i1 = 1
			j1 = 0
			k1 = 0
			i2 = 1
			j2 = 1
			k2 = 0
		} else if x0 >= z0 {
			i1 = 1
			j1 = 0
			k1 = 0
			i2 = 1
			j2 = 0
			k2 = 1
		} else {
			i1 = 0
			j1 = 0
			k1 = 1
			i2 = 1
			j2 = 0
			k2 = 1
		}
	} else { // x0<y0
		if y0 < z0 {
			i1 = 0
			j1 = 0
			k1 = 1
			i2 = 0
			j2 = 1
			k2 = 1
		} else if x0 < z0 {
			i1 = 0
			j1 = 1
			k1 = 0
			i2 = 0
			j2 = 1
			k2 = 1
		} else {
			i1 = 0
			j1 = 1
			k1 = 0
			i2 = 1
			j2 = 1
			k2 = 0
		}
	}

	// A step of (1,0,0) in (i,j,k) means a step of (1-c,-c,-c) in (x,y,z),
	// a step of (0,1,0) in (i,j,k) means a step of (-c,1-c,-c) in (x,y,z), and
	// a step of (0,0,1) in (i,j,k) means a step of (-c,-c,1-c) in (x,y,z), where
	// c = 1/6.
	x1 := x0 - float64(i1) + g3 // Offsets for second corner in (x,y,z) coords
	y1 := y0 - float64(j1) + g3
	z1 := z0 - float64(k1) + g3
	x2 := x0 - float64(i2) + 2.0*g3 // Offsets for third corner in (x,y,z) coords
	y2 := y0 - float64(j2) + 2.0*g3
	z2 := z0 - float64(k2) + 2.0*g3
	x3 := x0 - 1.0 + 3.0*g3 // Offsets for last corner in (x,y,z) coords
	y3 := y0 - 1.0 + 3.0*g3
	z3 := z0 - 1.0 + 3.0*g3

	// Work out the hashed gradient indices of the four simplex corners
	ii := int(i) & 255
	jj := int(j) & 255
	kk := int(k) & 255
	gi0 := perm[ii+int(perm[jj+int(perm[kk])])] % 12
	gi1 := perm[ii+i1+int(perm[jj+j1+int(perm[kk+k1])])] % 12
	gi2 := perm[ii+i2+int(perm[jj+j2+int(perm[kk+k2])])] % 12
	gi3 := perm[ii+1+int(perm[jj+1+int(perm[kk+1])])] % 12

	// Calculate the contribution from the four corners
	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0
	if t0 < 0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * dot(grad3[gi0], x0, y0, z0)
	}

	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1
	if t1 < 0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * dot(grad3[gi1], x1, y1, z1)
	}

	t2 := 0.6 - x2*x2 - y2*y2 - z2*z2
	if t2 < 0 {
		n2 = 0.0
	} else {
		t2 *= t2
		n2 = t2 * t2 * dot(grad3[gi2], x2, y2, z2)
	}

	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3
	if t3 < 0 {
		n3 = 0.0
	} else {
		t3 *= t3
		n3 = t3 * t3 * dot(grad3[gi3], x3, y3, z3)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to stay just inside [-1,1]
	return 32.0 * (n0 + n1 + n2 + n3)
}

func Noise2(xin, yin float64) float64 {
	var n0, n1, n2 float64 // Noise contributions from the three corners

	// Skew the input space to determine which simplex cell we're in
	s := (xin + yin) * f2 // Hairy factor for 2D
	i := math.Floor(xin + s)
	j := math.Floor(yin + s)
	t := (i + j) * g2
	X0 := i - t // Unskew the cell origin back to (x,y) space
	Y0 := j - t
	x0 := xin - X0 // The x,y distances from the cell origin
	y0 := yin - Y0

	// For the 2D case, the simplex shape is an equilateral triangle.

	// Determine which simplex we are in.
	var i1, j1 int // Offsets for second (middle) corner of simplex in (i,j) coords
	if x0 > y0 {
		i1 = 1
		j1 = 0
	} else {
		i1 = 0
		j1 = 1
	}

	// A step of (1,0) in (i,j) means a step of (1-c,-c) in (x,y), and
	// a step of (0,1) in (i,j) means a step of (-c,1-c) in (x,y), where
	// c = (3-sqrt(3))/6

	x1 := x0 - float64(i1) + g2 // Offsets for middle corner in (x,y) unskewed coords
	y1 := y0 - float64(j1) + g2
	x2 := x0 + g22 // Offsets for last corner in (x,y) unskewed coords
	y2 := y0 + g22

	// Work out the hashed gradient indices of the three simplex corners
	ii := int(i) & 255
	jj := int(j) & 255
	gi0 := perm[ii+int(perm[jj])] % 12
	gi1 := perm[ii+i1+int(perm[jj+j1])] % 12
	gi2 := perm[ii+1+int(perm[jj+1])] % 12

	// Calculate the contribution from the three corners
	t0 := 0.5 - x0*x0 - y0*y0
	if t0 < 0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * dot(grad3[gi0], x0, y0) // (x,y) of grad3 used for 2D gradient
	}

	t1 := 0.5 - x1*x1 - y1*y1
	if t1 < 0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * dot(grad3[gi1], x1, y1)
	}

	t2 := 0.5 - x2*x2 - y2*y2
	if t2 < 0 {
		n2 = 0.0
	} else {
		t2 *= t2
		n2 = t2 * t2 * dot(grad3[gi2], x2, y2)
	}

	// Add contributions from each corner to get the final noise value.
	// The result is scaled to return values in the interval [-1,1].
	return 70.0 * (n0 + n1 + n2)
}

func Noise4(x, y, z, w float64) float64 {
	var n0, n1, n2, n3, n4 float64 // Noise contributions from the five corners

	// Skew the (x,y,z,w) space to determine which cell of 24 simplices we're in
	s := (x + y + z + w) * f4 // Factor for 4D skewing
	i := math.Floor(x + s)
	j := math.Floor(y + s)
	k := math.Floor(z + s)
	l := math.Floor(w + s)

	t := (i + j + k + l) * g4 // Factor for 4D unskewing
	X0 := i - t               // Unskew the cell origin back to (x,y,z,w) space
	Y0 := j - t
	Z0 := k - t
	W0 := l - t
	x0 := x - X0 // The x,y,z,w distances from the cell origin
	y0 := y - Y0
	z0 := z - Z0
	w0 := w - W0

	// For the 4D case, the simplex is a 4D shape I won't even try to describe.
	// To find out which of the 24 possible simplices we're in, we need to
	// determine the magnitude ordering of x0, y0, z0 and w0.
	// The method below is a good way of finding the ordering of x,y,z,w and
	// then find the correct traversal order for the simplex we’re in.
	// First, six pair-wise comparisons are performed between each possible pair
	// of the four coordinates, and the results are used to add up binary bits
	// for an integer index.
	var c1, c2, c3, c4, c5, c6 int
	if x0 > y0 {
		c1 = 32
	}
	if x0 > z0 {
		c2 = 16
	}
	if y0 > z0 {
		c3 = 8
	}
	if x0 > w0 {
		c4 = 4
	}
	if y0 > w0 {
		c5 = 2
	}
	if z0 > w0 {
		c6 = 1
	}
	c := c1 + c2 + c3 + c4 + c5 + c6
	var i1, j1, k1, l1 int // The integer offsets for the second simplex corner
	var i2, j2, k2, l2 int // The integer offsets for the third simplex corner
	var i3, j3, k3, l3 int // The integer offsets for the fourth simplex corner

	// simplex[c] is a 4-vector with the numbers 0, 1, 2 and 3 in some order.
	// Many values of c will never occur, since e.g. x>y>z>w makes x<z, y<w and x<w
	// impossible. Only the 24 indices which have non-zero entries make any sense.
	// We use a thresholding to set the coordinates in turn from the largest magnitude.

	// The number 3 in the "simplex" array is at the position of the largest coordinate.
	if simplex[c][0] >= 3 {
		i1 = 1
	}
	if simplex[c][1] >= 3 {
		j1 = 1
	}
	if simplex[c][2] >= 3 {
		k1 = 1
	}
	if simplex[c][3] >= 3 {
		l1 = 1
	}

	// The number 2 in the "simplex" array is at the second largest coordinate.
	if simplex[c][0] >= 2 {
		i2 = 1
	}
	if simplex[c][1] >= 2 {
		j2 = 1
	}
	if simplex[c][2] >= 2 {
		k2 = 1
	}
	if simplex[c][3] >= 2 {
		l2 = 1
	}

	// The number 1 in the "simplex" array is at the second smallest coordinate.
	if simplex[c][0] >= 1 {
		i3 = 1
	}
	if simplex[c][1] >= 1 {
		j3 = 1
	}
	if simplex[c][2] >= 1 {
		k3 = 1
	}
	if simplex[c][3] >= 1 {
		l3 = 1
	}

	// The fifth corner has all coordinate offsets = 1, so no need to look that up.
	x1 := x0 - float64(i1) + g4 // Offsets for second corner in (x,y,z,w) coords
	y1 := y0 - float64(j1) + g4
	z1 := z0 - float64(k1) + g4
	w1 := w0 - float64(l1) + g4

	x2 := x0 - float64(i2) + g42 // Offsets for third corner in (x,y,z,w) coords
	y2 := y0 - float64(j2) + g42
	z2 := z0 - float64(k2) + g42
	w2 := w0 - float64(l2) + g42

	x3 := x0 - float64(i3) + g43 // Offsets for fourth corner in (x,y,z,w) coords
	y3 := y0 - float64(j3) + g43
	z3 := z0 - float64(k3) + g43
	w3 := w0 - float64(l3) + g43

	x4 := x0 + g44 // Offsets for last corner in (x,y,z,w) coords
	y4 := y0 + g44
	z4 := z0 + g44
	w4 := w0 + g44

	// Work out the hashed gradient indices of the five simplex corners
	ii := int(i) & 255
	jj := int(j) & 255
	kk := int(k) & 255
	ll := int(l) & 255

	gi0 := perm[ii+perm[jj+perm[kk+perm[ll]]]] % 32
	gi1 := perm[ii+i1+perm[jj+j1+perm[kk+k1+perm[ll+l1]]]] % 32
	gi2 := perm[ii+i2+perm[jj+j2+perm[kk+k2+perm[ll+l2]]]] % 32
	gi3 := perm[ii+i3+perm[jj+j3+perm[kk+k3+perm[ll+l3]]]] % 32
	gi4 := perm[ii+1+perm[jj+1+perm[kk+1+perm[ll+1]]]] % 32

	// Calculate the contribution from the five corners
	t0 := 0.6 - x0*x0 - y0*y0 - z0*z0 - w0*w0
	if t0 < 0 {
		n0 = 0.0
	} else {
		t0 *= t0
		n0 = t0 * t0 * dot(grad4[gi0], x0, y0, z0, w0)
	}

	t1 := 0.6 - x1*x1 - y1*y1 - z1*z1 - w1*w1
	if t1 < 0 {
		n1 = 0.0
	} else {
		t1 *= t1
		n1 = t1 * t1 * dot(grad4[gi1], x1, y1, z1, w1)
	}

	t2 := 0.6 - x2*x2 - y2*y2 - z2*z2 - w2*w2
	if t2 < 0 {
		n2 = 0.0
	} else {
		t2 *= t2
		n2 = t2 * t2 * dot(grad4[gi2], x2, y2, z2, w2)
	}

	t3 := 0.6 - x3*x3 - y3*y3 - z3*z3 - w3*w3
	if t3 < 0 {
		n3 = 0.0
	} else {
		t3 *= t3
		n3 = t3 * t3 * dot(grad4[gi3], x3, y3, z3, w3)
	}

	t4 := 0.6 - x4*x4 - y4*y4 - z4*z4 - w4*w4
	if t4 < 0 {
		n4 = 0.0
	} else {
		t4 *= t4
		n4 = t4 * t4 * dot(grad4[gi4], x4, y4, z4, w4)
	}

	// Sum up and scale the result to cover the range [-1,1]
	return 27.0 * (n0 + n1 + n2 + n3 + n4)
}
