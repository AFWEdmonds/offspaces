const content = document.getElementById("#content");
let offspaceData = []
let index = 0;
let adminKey = "";


async function queryOffspaces(append = false) {
    const response = await fetch("http://localhost:3333/?" + formToQuery('form'));
    const { data, total } = await response.json();

    totalCount = total;

    if (!append) {
        document.getElementById("content").innerHTML = "";
        index = 0;
        offspaceData = [];
    }

    const frag = document.createDocumentFragment();

    data.forEach((v, i) => {
        offspaceData.push(v);
        renderCard(frag, v, false, offspaceData.length - 1);
    });

    document.getElementById("content").appendChild(frag);

    // show/hide load more button
    const loadMoreBtn = document.getElementById("loadMoreBtn");
    if (offspaceData.length >= totalCount) {
        loadMoreBtn.style.display = "none";
    } else {
        loadMoreBtn.style.display = "block";
    }
}

function formToQuery(formName) {
    const form = document.getElementById(formName);
    const formData = new FormData(form);
    let key = adminKey ? '&adminKey=' + adminKey : '';
    return new URLSearchParams(formData).toString() + "&index=" + index + key;
}


function renderCard(frag, v, admin, index) {
    const card = document.createElement('div');
    card.className = 'card my-4';

    const img = document.createElement('img');
    img.className = 'card-img-top max-vh-50';
    img.alt = 'The offspace';
    img.src = v.photo;
    card.appendChild(img);

    const cardBody = document.createElement('div');
    cardBody.className = 'card-body';

    const cardTitle = document.createElement('h5');
    cardTitle.className = 'card-title';
    cardTitle.innerHTML = v.name;
    if (!v.published) {
        v.name += '(Unpublished)';
    }
    cardBody.appendChild(cardTitle);

    const cardText = document.createElement('p');
    cardText.className = 'card-text';
    cardText.innerHTML = `${v.street}<br>${v.postcode} ${v.city}`;
    cardBody.appendChild(cardText);

    // ---- Opening Times ----
    if (v.opening_times) {
        let ot;
        try {
            // v.OpeningTimes may already be JSON or a parsed object
            ot = typeof v.opening_times === "string"
                ? JSON.parse(v.opening_times)
                : v.opening_times;
        } catch (e) {
            ot = null;
        }

        if (ot) {
            const days = ["mon","tue","wed","thu","fri","sat","sun"];
            const pretty = {
                mon: "Mon",
                tue: "Tue",
                wed: "Wed",
                thu: "Thu",
                fri: "Fri",
                sat: "Sat",
                sun: "Sun"
            };

            let output = "<strong>Opening times</strong><br>";

            days.forEach(d => {
                const day = ot[d];
                if (day.start) {
                    const start = day.start || "";
                    const end   = day.end   || "";
                    if (start && end) {
                        output += `${pretty[d]}: ${start}–${end}<br>`;
                    } else {
                        output += `${pretty[d]}: —<br>`;
                    }
                }
            });

            const otP = document.createElement('p');
            otP.className = "card-text small";
            otP.innerHTML = output;
            cardBody.appendChild(otP);
        }
    }

    const visitWebsite = document.createElement('a');
    visitWebsite.href = v.Website;
    visitWebsite.className = 'card-link';
    visitWebsite.innerHTML = "Website";
    cardBody.appendChild(visitWebsite);

    const visitSM = document.createElement('a');
    visitSM.href = v.SocialMedia;
    visitSM.className = 'card-link';
    visitSM.innerHTML = "Social media";
    cardBody.appendChild(visitSM);

    if (admin) {
        cardBody.appendChild(document.createElement('br'));
        const edit = document.createElement('button');
        edit.className = 'btn btn-warning';
        edit.innerHTML = "Admin Edit";
        edit.type = 'button';
        edit.dataset.index = index; // string
        edit.addEventListener('click', (e) => {
            const idx = Number(e.currentTarget.dataset.index); // parse back to number
            console.log(idx);
            editOffspace(offspaceData.data[idx]);
        });        cardBody.appendChild(edit);
    }

    card.appendChild(cardBody);
    frag.appendChild(card);
}


// Generate HH:MM values every 30 min
function generateTimes() {
    const result = [];
    for (let h = 0; h < 24; h++) {
        for (const m of ["00", "30"]) {
            result.push(`${String(h).padStart(2, "0")}:${m}`);
        }
    }
    return result;
}

const times = generateTimes();

// Populate all selects
document.querySelectorAll(".time-start, .time-end").forEach(select => {
    select.innerHTML = '<option value=""></option>'; // blank option
    times.forEach(t => {
        const opt = document.createElement("option");
        opt.value = t;
        opt.textContent = t;
        select.appendChild(opt);
    });
});

// Enable/disable logic
document.querySelectorAll(".closed-check").forEach(check => {
    check.addEventListener("change", () => {
        const row = check.closest(".row");
        const start = row.querySelector(".time-start");
        const end = row.querySelector(".time-end");

        const closed = check.checked;

        start.disabled = closed;
        end.disabled = closed;

        if (closed) {
            start.value = "";
            end.value = "";
        }
    });

    // Trigger initial state
    check.dispatchEvent(new Event("change"));
});

queryOffspaces();

document.getElementById("form").addEventListener("submit", (e) => {
    e.preventDefault(); // stop the browser from reloading the page
    queryOffspaces();
});
