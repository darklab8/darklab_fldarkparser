package trades

/*
Game graph simplifies for us conversion of data from Freelancer space simulator to different graph algorithms.
*/

import (
	"math"
	"reflect"
)

type VertexName string

type GameGraph struct {
	matrix                  map[VertexName]map[VertexName]float64
	IndexByNick             map[VertexName]int
	NicknameByIndex         map[int]VertexName
	AllowedVertixesForCalcs map[VertexName]bool // Consider deleting this
	AvgCruiseSpeed          int
	idsNamesByNick          map[VertexName]int
	IsTradelane             map[VertexName]bool
}

func NewGameGraph(avgCruiseSpeed int) *GameGraph {
	return &GameGraph{
		matrix:                  make(map[VertexName]map[VertexName]float64),
		IndexByNick:             map[VertexName]int{},
		NicknameByIndex:         make(map[int]VertexName),
		AllowedVertixesForCalcs: make(map[VertexName]bool),
		AvgCruiseSpeed:          avgCruiseSpeed,
		idsNamesByNick:          make(map[VertexName]int),
		IsTradelane:             make(map[VertexName]bool),
	}
}

func (f *GameGraph) SetEdge(keya string, keyb string, distance float64) {
	vertex, vertex_exists := f.matrix[VertexName(keya)]
	if !vertex_exists {
		vertex = make(map[VertexName]float64)
		f.matrix[VertexName(keya)] = vertex
	}

	if _, vert_target_exists := f.matrix[VertexName(keyb)]; !vert_target_exists {
		f.matrix[VertexName(keyb)] = make(map[VertexName]float64)
	}

	_, already_set := vertex[VertexName(keyb)]
	if already_set {
		return // otherwise u will overwrite tradelane distances.
	}

	vertex[VertexName(keyb)] = distance
}

func (f *GameGraph) SetIdsName(keya string, ids_name int) {
	f.idsNamesByNick[VertexName(keya)] = ids_name
}

func (f *GameGraph) SetIstRadelane(keya string) {
	f.IsTradelane[VertexName(keya)] = true
}

func GetDist(f *GameGraph, dist [][]int, keya string, keyb string) int {
	sourse_index, source_found := f.IndexByNick[VertexName(keya)]
	target_index, target_found := f.IndexByNick[VertexName(keyb)]
	_ = source_found
	if !source_found || !target_found {
		return INF
	}
	return dist[sourse_index][target_index]
}

type Path struct {
	Node     int
	NextNode int
	Dist     int
}

func GetPath(graph *GameGraph, parents [][]int, dist [][]int, source_key string, target_key string) []Path {
	// fmt.Println("get_path", source_key, target_key)
	S := []Path{}
	u, found_u := graph.IndexByNick[VertexName(target_key)] // target
	if !found_u {
		return []Path{}
	}
	_ = found_u
	source := graph.IndexByNick[VertexName(source_key)]

	add_node := func(u int) {
		path_to_add := Path{
			Node: u,
		}
		if len(S) > 0 {
			path_to_add.NextNode = S[len(S)-1].Node
		} else {
			path_to_add.NextNode = NO_PARENT
		}
		if path_to_add.Node != NO_PARENT && path_to_add.NextNode != NO_PARENT {
			path_to_add.Dist = dist[path_to_add.Node][path_to_add.NextNode]
		}

		S = append(S, path_to_add)
	}
	add_node(u)

	if parents[source][u] != NO_PARENT || u == source {
		for {
			u = parents[source][u]

			nickname := graph.NicknameByIndex[u]
			if _, ok := graph.IsTradelane[nickname]; ok {
				continue
			}

			add_node(u)

			if u == NO_PARENT {
				break
			}
		}
	}
	ReverseSlice(S)
	return S
}

// panic if s is not a slice
func ReverseSlice(s interface{}) {
	size := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, size-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}

type DetailedPath struct {
	PrevName    string
	NextName    string
	PrevIdsName int
	NextIdsName int
	PrevNode    int
	NextNode    int
	Dist        int
	TimeMinutes int
	TimeSeconds int
}

func (graph *GameGraph) GetPaths(parents [][]int, dist [][]int, source_key string, target_key string) []DetailedPath {
	var detailed_paths []DetailedPath

	paths := GetPath(graph, parents, dist, source_key, target_key)
	for _, path := range paths {
		minutes := int(math.Floor(float64(graph.GetTimeForDist(float64(path.Dist))) / 60))
		detailed_path := DetailedPath{
			PrevName:    string(graph.NicknameByIndex[path.Node]),
			NextName:    string(graph.NicknameByIndex[path.NextNode]),
			PrevIdsName: graph.idsNamesByNick[graph.NicknameByIndex[path.Node]],
			NextIdsName: graph.idsNamesByNick[graph.NicknameByIndex[path.NextNode]],
			PrevNode:    path.Node,
			NextNode:    path.NextNode,
			Dist:        path.Dist,
			TimeMinutes: int(minutes),
			TimeSeconds: graph.GetTimeForDist(float64(path.Dist)) - minutes*60,
		}

		detailed_paths = append(detailed_paths, detailed_path)
	}

	return detailed_paths
}
