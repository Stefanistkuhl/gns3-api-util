const classNameInput = document.getElementById('className');
const groupNameInput = document.getElementById('groupName');
const groupNumberInput = document.getElementById('groupNumber');
const selectedGroupSelect = document.getElementById('selectedGroup');
const fullNameInput = document.getElementById('fullName');
const userNameInput = document.getElementById('userName');
const emailInput = document.getElementById('email');
const passwordLengthInput = document.getElementById('passwordLength');
const passwordErrorP = document.getElementById('passwordError');
const addGroupBtn = document.getElementById('addGroupBtn');
const addStudentBtn = document.getElementById('addStudentBtn');
const generateJSONBtn = document.getElementById('generateJSONBtn');
const saveJSONBtn = document.getElementById('saveJSONBtn');

let groups = {};
let selectedGroup = '';

function generatePassword(length) {
	const clampedLength = Math.max(8, Math.min(length, 128));
	const charset = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()';
	return Array.from(crypto.getRandomValues(new Uint8Array(clampedLength)))
		.map(n => charset[n % charset.length])
		.join('');
}

function updateGroupSelect() {
	while (selectedGroupSelect.options.length > 1) {
		selectedGroupSelect.remove(1);
	}
	Object.keys(groups).forEach(key => {
		const opt = document.createElement('option');
		opt.value = key;
		opt.textContent = key;
		selectedGroupSelect.appendChild(opt);
	});
	if (selectedGroup && groups[selectedGroup]) {
		selectedGroupSelect.value = selectedGroup;
	} else {
		selectedGroupSelect.value = '';
		selectedGroup = '';
	}
}

addGroupBtn.addEventListener('click', () => {
	const className = classNameInput.value.trim();
	const groupName = groupNameInput.value.trim();
	const groupNumber = groupNumberInput.value.trim();
	if (!className || !groupName || !groupNumber) return;
	const key = `${className}-${groupName}-${groupNumber}`;
	if (!groups[key]) {
		groups[key] = { students: [] };
	}
	selectedGroup = key;
	groupNameInput.value = '';
	groupNumberInput.value = '';
	updateGroupSelect();
	selectedGroupSelect.value = selectedGroup;
});

selectedGroupSelect.addEventListener('change', () => {
	selectedGroup = selectedGroupSelect.value;
});

addStudentBtn.addEventListener('click', () => {
	const passwordLength = parseInt(passwordLengthInput.value, 10);
	if (passwordLength > 128) {
		passwordErrorP.textContent = 'Password length cannot exceed 128 characters.';
		passwordErrorP.style.display = '';
		return;
	} else {
		passwordErrorP.textContent = '';
		passwordErrorP.style.display = 'none';
	}
	if (!selectedGroup || !groups[selectedGroup]) return;
	const fullName = fullNameInput.value.trim();
	const userName = userNameInput.value.trim();
	const email = emailInput.value.trim();
	if (!fullName || !userName || !email) return;
	groups[selectedGroup].students.push({
		fullName,
		userName,
		password: generatePassword(passwordLength),
		email
	});
	fullNameInput.value = '';
	userNameInput.value = '';
	emailInput.value = '';
});

saveJSONBtn.addEventListener('click', () => {
	const className = classNameInput.value.trim();
	if (!className) return;
	const output = {
		[className]: groups
	};
	const blob = new Blob([JSON.stringify(output, null, 2)], { type: 'application/json' });
	const link = document.createElement('a');
	link.href = URL.createObjectURL(blob);
	link.download = `${className}.json`;
	link.click();
});

generateJSONBtn.addEventListener('click', async () => {
	const className = classNameInput.value.trim();
	if (!className) return;
	const output = {
		[className]: groups
	};
	const payload = JSON.stringify(output, null, 2)
	console.log(payload)
	const result = await sendData(payload)


});

async function sendData(payload) {
	const url = "http://localhost:8080/data";
	const options = {
		method: "POST",
		body: payload,
		headers: {
			'Content-Type': 'application/json'
		}
	};

	try {
		const response = await fetch(url, options);
		if (!response.ok) {
			throw new Error(`Response status: ${response.status}`);
		}

		const json = await response.json();
		console.log(json);
		return json;
	} catch (error) {
		console.error(error.message);
		throw error;
	}
}
