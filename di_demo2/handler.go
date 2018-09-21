package di_demo2

// import "net/http"

// // pretend test
// type memPostService struct {
// 	db map[int]*Post
// }
// func TestHandler(t *testing.T) {
// 	h := Handler{
// 		PostService: memPostService,
// 	}
// }

// type Handler struct {
// 	PaymentService interface {
// 		Charge()
// 	}
// 	PostService interface {
// 		Find(id int) (*Post, error)
// 		Update(*Post) error
// 		Publish(*Post) error
// 	}
// }

// func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
// 	// Get post ID & updated content from request
// 	id := parseID(r)
// 	content := parseContent(r)
// 	post, err := h.PostService.Find(id)
// 	if err != nil {
// 		// ...
// 	}
// 	post.Content = content
// 	err = h.PostService.Update(post)
// 	// ...
// }

// func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
// 	// Get post ID from request
// 	id := parseID(r)
// 	post, err := h.PostService.Find(id)
// 	if err != nil {
// 		// ...
// 	}
// 	err = h.PostService.Publish(post)
// 	// ...
// }
