package schemas

type Student struct {
	FullName *string `json:"fullName,omitempty"`
	UserName string  `json:"userName" validate:"required"`
	Password string  `json:"password" validate:"required,min=8,max=100"`
	Email    *string `json:"email,omitempty"`
}

type Group struct {
	Name     string    `json:"name" validate:"required"`
	Students []Student `json:"students" validate:"required"`
}

type Class struct {
	Name   string  `json:"name" validate:"required"`
	Desc   string  `json:"description" validate:"omitempty"`
	Groups []Group `json:"groups" validate:"required"`
}

type GroupListElement struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Name   string `json:"name"`
}

type Exercise struct {
	Name        string `json:"name"`
	ID          string `json:"id"`
	ClassName   string `json:"class_name"`
	GroupNumber string `json:"group_number"`
}

type SelectedExercise struct {
	ExerciseName string  `json:"exercise_name"`
	ClassName    *string `json:"class_name,omitempty"`
	GroupNumber  *string `json:"group_number,omitempty"`
}

type ExercisePool struct {
	PoolID      string `json:"pool_id"`
	ClassName   string `json:"class_name"`
	GroupNumber string `json:"group_number"`
}
