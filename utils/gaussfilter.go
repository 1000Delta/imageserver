package utils

import (
	"image"
	"image/color"
	"math"
	"sync"
)

func gaussFuncDD(x, y int, sigma float64) float64 {
	return 1 / (2 * math.Pi * sigma * sigma) * math.Exp(-float64(x*x+y*y)/(2*sigma*sigma))
}

// 对图像进行高斯模糊, radius是卷积核半径, 为 1 时只包含自身; routines 为线程数, 一半采用硬件线程数较好; sigma 为高斯方程权重
func GaussFuzzy(img image.Image, radius, routines int, sigma float64) image.Image {
	//var coreLen = radius*2 - 1
	b := img.Bounds()
	gaussWeights := make([][]float64, radius) // 坐标数
	// 计算高斯函数值
	gaussSum := float64(0)
	for x := 0; x < radius; x++ {
		gaussWeights[x] = make([]float64, radius)
		for y := 0; y <= x; y++ {
			gaussWeights[x][y] = gaussFuncDD(x, y, sigma)
			gaussSum += gaussWeights[x][y]
			if x != 0 || y != 0 {
				gaussSum += gaussWeights[x][y] * 3 // 坐标点中心对称 90deg
			}
			// 减少重复计算
			if x != y { // 排除对角线
				gaussWeights[y][x] = gaussWeights[x][y]
				if x != 0 && y != 0 { // 排除中线
					gaussSum += gaussWeights[y][x] * 4
				}
			}
		}
	}
	// 归一化
	for x := 0; x < radius; x++ {
		for y := 0; y <= x; y++ {
			gaussWeights[x][y] = gaussWeights[x][y] / gaussSum
			if x != y {
				gaussWeights[y][x] = gaussWeights[x][y]
			}
		}
	}
	// TODO 多线程计算
	wg := sync.WaitGroup{}
	wg.Add(routines)
	xSegment := b.Dx() / routines
	nImg := image.NewRGBA64(b)
	for c := 0; c < routines; c++ {
		subNewImg := nImg.SubImage(image.Rect(b.Min.X+xSegment*c, b.Min.Y, b.Max.X+xSegment*(c+1), b.Max.Y)).(*image.RGBA64)
		go func(in image.Image, out *image.RGBA64, segN int) {
			defer wg.Done()
			var xBegin = xSegment * segN
			b := out.Bounds()
			for x := b.Min.X; x < b.Max.X; x++ {
				for y := b.Min.Y; y < b.Max.Y; y++ {
					var rgba [4]uint32
					//cs := make([]color.Color, 0)
					// 截取边缘值 边界点采用镜像值
					for i := xBegin + x - radius + 1; i < xBegin+x+radius; i++ {
						for j := y - radius + 1; j < y+radius; j++ {
							var r, g, b, a uint32
							switch {
							case i < 0 && j < 0:
								//cs = append(cs, in.At(-i, -j))
								r, g, b, a = in.At(-i, -j).RGBA()
							case i < 0:
								//cs = append(cs, in.At(-i, j))
								r, g, b, a = in.At(-i, j).RGBA()
							case j < 0:
								//cs = append(cs, in.At(i, -j))
								r, g, b, a = in.At(i, -j).RGBA()
							default:
								//cs = append(cs, in.At(i, j))
								r, g, b, a = in.At(i, j).RGBA()
							}
							gx := i - x - xBegin
							gy := j - y
							//fmt.Printf("(%d,%d)\t(%d, %d)\n", i, j, gx, gy)
							if gx < 0 {
								gx = -gx
							}
							if gy < 0 {
								gy = -gy
							}
							rgba[0] += uint32(float64(r) * gaussWeights[gx][gy])
							rgba[1] += uint32(float64(g) * gaussWeights[gx][gy])
							rgba[2] += uint32(float64(b) * gaussWeights[gx][gy])
							rgba[3] += uint32(float64(a) * gaussWeights[gx][gy])
						}
					}
					//for i, c := range cs {
					//	cx, cy := i%coreLen-radius+1, i/coreLen-radius+1
					//	if cx < 0 {
					//		cx = -cx
					//	}
					//	if cy < 0 {
					//		cy = -cy
					//	}
					//	r, g, b, a := c.RGBA()
					//	rgba[0] += uint32(float64(r) * gaussWeights[cx][cy])
					//	rgba[1] += uint32(float64(g) * gaussWeights[cx][cy])
					//	rgba[2] += uint32(float64(b) * gaussWeights[cx][cy])
					//	rgba[3] += uint32(float64(a) * gaussWeights[cx][cy])
					//}
					nc := &color.RGBA64{
						R: uint16(rgba[0]),
						G: uint16(rgba[1]),
						B: uint16(rgba[2]),
						A: uint16(rgba[3]),
					}
					out.Set(x, y, nc)
				}
			}
		}(img, subNewImg, c)
	}
	wg.Wait()
	return nImg
}
