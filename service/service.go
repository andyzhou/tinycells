package service

/**
 * Inter service face
 * @author <AndyZhou>
 * @mail <diudiu8848@163.com>
 */

 //inter service info
 type InterService struct {
 	httpPort int
 	http *HttpService
 	redis *RedisService
 	mysql *MysqlService
 }

 //construct
func NewInterService() *InterService {
	//self init
	this := &InterService{}
	return  this
}

//create relate service
func (s *InterService) CreateHttpService(queues int) *HttpService {
	if s.http != nil {
		return s.http
	}
	s.http = NewHttpService(queues)
	return s.http
}

func (s *InterService) CreateRedisService() *RedisService {
	if s.redis != nil {
		return s.redis
	}
	s.redis = NewRedisService()
	return s.redis
}

func (s *InterService) CreateMysqlService() *MysqlService {
	if s.mysql != nil {
		return s.mysql
	}
	s.mysql = NewMysqlService()
	return s.mysql
}

//get relate service
func (s *InterService) GetHttpService() *HttpService {
	return s.http
}

func (s *InterService) GetRedisService() *RedisService {
	return s.redis
}

func (s *InterService) GetMysqlService() *MysqlService {
	return s.mysql
}

//quit
func (s *InterService) Quit() {
	s.redis.Quit()
	s.mysql.Quit()
	s.http.Quit()
}
