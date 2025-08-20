const classNameInput = document.getElementById("className");
const groupNameInput = document.getElementById("groupName");
const groupNumberInput = document.getElementById("groupNumber");
const selectedGroupSelect = document.getElementById("selectedGroup");
const fullNameInput = document.getElementById("fullName");
const userNameInput = document.getElementById("userName");
const emailInput = document.getElementById("email");
const passwordLengthInput = document.getElementById("passwordLength");
const passwordErrorP = document.getElementById("passwordError");
const userDataErrorP = document.getElementById("userDataError");
const addGroupBtn = document.getElementById("addGroupBtn");
const addStudentBtn = document.getElementById("addStudentBtn");
const generateJSONBtn = document.getElementById("generateJSONBtn");
const saveJSONBtn = document.getElementById("saveJSONBtn");
const clearJSONBtn = document.getElementById("clearJSONBtn");
const themeToggle = document.getElementById("themeToggle");
const body = document.body;

let classData = {
	name: "",
	groups: [],
};
let selectedGroup = null;
let groupNumberCount = 1;

// Check for saved theme preference
const savedTheme = localStorage.getItem("theme");
if (savedTheme === "dark") {
	body.classList.add("dark-mode");
}

function generatePassword(length) {
	const clampedLength = Math.max(8, Math.min(length, 128));
	const charset =
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()";
	return Array.from(crypto.getRandomValues(new Uint8Array(clampedLength)))
		.map((n) => charset[n % charset.length])
		.join("");
}

function updateGroupSelect() {
	while (selectedGroupSelect.options.length > 1) {
		selectedGroupSelect.remove(1);
	}
	classData.groups.forEach((group) => {
		const opt = document.createElement("option");
		opt.value = group.name;
		opt.textContent = group.name;
		selectedGroupSelect.appendChild(opt);
	});
	if (selectedGroup && classData.groups.includes(selectedGroup)) {
		selectedGroupSelect.value = selectedGroup.name;
	} else {
		selectedGroupSelect.value = "";
		selectedGroup = null;
	}
}

addGroupBtn.addEventListener("click", () => {
	const className = classNameInput.value.trim();
	const groupName = groupNameInput.value.trim();
	let groupNumber = groupNumberInput.value.trim();

	if (!className || !groupName) return;

	if (!classData.name) {
		classData.name = className;
	} else if (classData.name !== className) {
		// If class name changes, clear existing groups or handle as an error
		alert(
			"You can only have one class name per JSON. Please clear the existing data to create a new class."
		);
		return;
	}

	if (!groupNumber) {
		groupNumber = groupNumberCount;
		groupNumberCount++;
	}

	let finalGroupName = `${className}-${groupName}-${groupNumber}`;
	let i = 0;
	while (checkGroupNumberReuse(finalGroupName, classData.groups)) {
		finalGroupName = `${className}-${groupName}-${i}`;
		i++;
	}

	const newGroup = { name: finalGroupName, students: [] };
	classData.groups.push(newGroup);
	selectedGroup = newGroup;
	groupNameInput.value = "";
	groupNumberInput.value = "";
	updateGroupSelect();
	selectedGroupSelect.value = selectedGroup.name;
});

selectedGroupSelect.addEventListener("change", () => {
	const groupName = selectedGroupSelect.value;
	selectedGroup = classData.groups.find((group) => group.name === groupName);
});

addStudentBtn.addEventListener("click", () => {
	const re = new RegExp("^[\\w\\-\\.]+@([\\w-]+\\.)+[\\w-]{2,}$", "gm");
	const passwordLength = parseInt(passwordLengthInput.value, 10);

	if (passwordLength > 128) {
		passwordErrorP.textContent = "Password length cannot exceed 128 characters.";
		passwordErrorP.style.display = "";
		return;
	} else if (passwordLength < 8) {
		passwordErrorP.textContent = "Password length cannot be bellow 8 characters.";
		passwordErrorP.style.display = "";
		return;
	} else {
		passwordErrorP.textContent = "";
		passwordErrorP.style.display = "none";
	}

	if (!selectedGroup) return;

	const fullName = fullNameInput.value.trim();
	const userNameRaw = userNameInput.value.trim();

	if (userNameRaw.length < 3) {
		userDataErrorP.textContent = "Username cannot be shorter than 3 characters.";
		userDataErrorP.style.display = "";
		return;
	} else {
		userDataErrorP.textContent = "";
		userDataErrorP.style.display = "none";
	}

	let userName = userNameRaw;
	let i = 0;
	while (checkUserNameReuse(userName, classData.groups)) {
		userName = `${userNameRaw}-${i}`;
		i++;
	}

	const emailRaw = emailInput.value.trim();
	if (!re.test(emailRaw)) {
		userDataErrorP.textContent = "Please enter a valid email address";
		userDataErrorP.style.display = "";
		return;
	} else {
		userDataErrorP.textContent = "";
		userDataErrorP.style.display = "none";
	}

	if (checkEmailReuse(emailRaw, classData.groups)) {
		userDataErrorP.textContent = "This email has been reused, please use a different one";
		userDataErrorP.style.display = "";
		return;
	} else {
		userDataErrorP.textContent = "";
		userDataErrorP.style.display = "none";
	}

	if (!fullName || !userName || !emailRaw) return;

	selectedGroup.students.push({
		fullName,
		userName,
		password: generatePassword(passwordLength),
		email: emailRaw,
	});

	fullNameInput.value = "";
	userNameInput.value = "";
	emailInput.value = "";
});

saveJSONBtn.addEventListener("click", () => {
	if (!classData.name) return;
	const blob = new Blob([JSON.stringify(classData, null, 2)], {
		type: "application/json",
	});
	const link = document.createElement("a");
	link.href = URL.createObjectURL(blob);
	link.download = `${classData.name}.json`;
	link.click();
});

generateJSONBtn.addEventListener("click", async () => {
	if (!classData.name) return;
	const payload = JSON.stringify(classData, null, 2);
	const result = await sendData(payload);
	console.log(result); // Or display the result on the page
});

clearJSONBtn.addEventListener("click", async () => {
	classData = {
		name: "",
		groups: [],
	};
	selectedGroup = null;
	groupNumberCount = 1;
	updateGroupSelect();
	classNameInput.value = "";
	groupNameInput.value = ""; // Clear group name input as well
	groupNumberInput.value = "";
	selectedGroupSelect.value = "";
	fullNameInput.value = "";
	userNameInput.value = "";
	emailInput.value = "";
	passwordLengthInput.value = 12;
	passwordErrorP.textContent = "";
	passwordErrorP.style.display = "none";
	userDataErrorP.textContent = "";
	userDataErrorP.style.display = "none";
});

async function sendData(payload) {
	const url = "http://localhost:8080/data";
	const options = {
		method: "POST",
		body: payload,
		headers: {
			"Content-Type": "application/json",
		},
	};

	try {
		const response = await fetch(url, options);
		if (!response.ok) {
			throw new Error(`Response status: ${response.status} `);
		}

		const json = await response.json();
		return json;
	} catch (error) {
		console.error(error.message);
		throw error;
	}
}

function checkUserNameReuse(name, groupsArray) {
	for (const group of groupsArray) {
		for (const student of group.students) {
			if (student.userName === name) {
				return true;
			}
		}
	}
	return false;
}

function checkEmailReuse(email, groupsArray) {
	for (const group of groupsArray) {
		for (const student of group.students) {
			if (student.email === email) {
				return true;
			}
		}
	}
	return false;
}

function checkGroupNumberReuse(groupname, groupsArray) {
	return groupsArray.some((group) => group.name === groupname);
}

document.addEventListener("DOMContentLoaded", () => {
	// Function to set a cookie
	function setCookie(name, value, days) {
		const expires = new Date();
		expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
		document.cookie = `${name}=${value};expires=${expires.toUTCString()};path=/`;
	}

	// Function to get a cookie value
	function getCookie(name) {
		const nameEQ = name + "=";
		const ca = document.cookie.split(";");
		for (let i = 0; i < ca.length; i++) {
			let c = ca[i];
			while (c.charAt(0) === " ") c = c.substring(1, c.length);
			if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
		}
		return null;
	}

	// Check if there's a saved theme preference in cookies
	const savedTheme = getCookie("theme");
	if (savedTheme === "dark") {
		body.classList.add("dark-mode");
	}

	// Theme toggle click handler
	themeToggle.addEventListener("click", () => {
		body.classList.toggle("dark-mode");

		// Save the current theme preference in a cookie that expires in 365 days
		if (body.classList.contains("dark-mode")) {
			setCookie("theme", "dark", 365);
		} else {
			setCookie("theme", "light", 365);
		}
	});
});
