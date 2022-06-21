package grpc_pool

type stack struct {
	values []*rpcConn
}

func (s *stack) Peek() (conn *rpcConn) {
	if len(s.values) == 0 {
		return nil
	}
	return s.values[len(s.values)-1]
}

func (s *stack) Push(conn *rpcConn) {
	s.values = append(s.values, conn)
}

func (s *stack) Pop() (conn *rpcConn) {
	conn = s.values[len(s.values)-1]
	s.values = s.values[:len(s.values)-1]
	return
}

func (s *stack) Remove(id string) (conn *rpcConn) {
	for i, v := range s.values {
		if v.id == id {
			conn = v
			if i == len(s.values)-1 {
				s.values = s.values[:len(s.values)-1]
			} else {
				s.values = append(s.values[:i], s.values[i+1:]...)
			}
			break
		}
	}
	return
}

func (s *stack) Empty() bool {
	return len(s.values) == 0
}

func (s *stack) Size() int {
	return len(s.values)
}
