package transforms

import (
	"github.com/dustismo/heavyfishdesign/path"
)

type ReorderTransform struct {
	Precision int
}

// checks all the paths in pList to see which should be next.
// Note this is a linear search, so this is fully N^2
func (st ReorderTransform) doPath(p path.Path, pList []path.Path) (path.Path, []path.Path, error) {
	// based on p, this will return the next closest
	// path from pList
	_, endPoint := path.GetStartAndEnd(path.TrimTailMove(p.Segments()))

	// distances organized by index
	sDistances := make([]float64, len(pList))
	eDistances := make([]float64, len(pList))
	shortestSInd := 0
	shortestEInd := 0
	for i, pth := range pList {
		s, e := path.GetStartAndEnd(pth.Segments())
		sDistances[i] = path.Distance(endPoint, s)
		eDistances[i] = path.Distance(endPoint, e)
		if sDistances[i] < sDistances[shortestSInd] {
			shortestSInd = i
		}

		if eDistances[i] < eDistances[shortestEInd] {
			shortestEInd = i
		}
	}

	// TODO: we should look at the end list as well
	// if eDistance is the shortest, we should
	// reverse it and return it.
	i := shortestSInd
	distance := sDistances[i]

	retPath := pList[i]
	if eDistances[shortestEInd] < sDistances[i] {
		// reverse and reset
		i = shortestEInd
		distance = eDistances[i]
		retPath = pList[i]
		r, err := PathReverse{}.PathTransform(retPath)
		if err != nil {
			return p, pList, err
		}
		retPath = r
	}

	// is the distance close enough?
	if !path.PrecisionEquals(distance, 0, st.Precision) {
		i = 0
		retPath = pList[0]
	}

	// remove the item from the list
	newList := append(pList[:i], pList[i+1:]...)
	return path.NewPathFromSegments(append(p.Segments(), retPath.Segments()...)), newList, nil
}

// Transform that will reorder any sections that
// start and end at the same place
func (st ReorderTransform) PathTransform(p path.Path) (path.Path, error) {
	paths := path.SplitPathOnMove(p)
	if len(paths) <= 1 {
		// nothing to do here!
		return p, nil
	}
	retList := []path.Segment{}
	retList = append(retList, paths[0].Segments()...)
	pList := paths[1:]
	pth := paths[0]
	for len(pList) > 0 {
		p1, pL, err := st.doPath(pth, pList)
		if err != nil {
			return p, err
		}

		pList = pL
		retList = append(retList, p1.Segments()...)
		pth = p1
	}
	return DedupSegmentsTransform{Precision: st.Precision}.PathTransform(pth)
}
