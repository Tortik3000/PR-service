import http from 'k6/http';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE = __ENV.BASE_URL || `http://localhost:8080`;
const hdrs = { headers: { 'Content-Type': 'application/json' } };

function uuid() {
  return Math.random().toString(36).substring(2);
}

export const options = {
    scenarios: {
        add_team: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'addTeam',
        },
        create_pr: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'createPR',
        },
         merge_pr: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'mergePR',
        },
        get_team_info: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'getTeamInfo',
        },
        get_user_prs: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'getUserPRs',
        },
        set_is_active: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'setIsActive',
        },
    },

};

export function setup() {
    const teams = [];
    const users = [];

    for (let i = 0; i < 10; i++) {
        const team_name = `Team-${uuid()}`;
        const members = [];

        for (let j = 0; j < 10; j++) {
            const user_id = uuid();
            const username = `User-${uuid()}`;
            const is_active = true;

            members.push({ user_id, username, is_active });
            users.push(user_id);
        }

        const res = http.post(`${BASE}/team/add`,
            JSON.stringify({ team_name, members }), hdrs);
        if (res.status !== 201) {
            console.error(`Ошибка при addTeam ${team_name}: ${res.status}`);
        }

        teams.push(team_name);
    }

    const prs = [];
    users.forEach((user_id, idx) => {
        const pr_id = uuid();
        const pr_name = `Feature-${uuid()}`;

        const res = http.post(
            `${BASE}/pullRequest/create`,
            JSON.stringify({ pull_request_id: pr_id, pull_request_name: pr_name, author_id: user_id }),
            hdrs
        );

        if (res.status === 201) {
            prs.push(pr_id);
        } else {
            console.error(`Ошибка при prCreate ${pr_id}: ${res.status}`);
        }
    });

    return { teams, users, prs };
}

export function addTeam() {
    const team_name = `Team-${uuid()}`;
    const members = [
        { user_id: uuid(), username: `User-${uuid()}`, is_active: true },
        { user_id: uuid(), username: `User-${uuid()}`, is_active: true },
    ];

    const res = http.post(`${BASE}/team/add`,
        JSON.stringify({ team_name, members }), hdrs);
    if (res.status !== 201) {
        console.error(`Ошибка при addTeam ${team_name}: ${res.status}`);
    }
}

export function createPR(data) {
    const author_id = randomItem(data.users);
    const pr_id = uuid();
    const pr_name = `Feature-${uuid()}`;

    const res = http.post(`${BASE}/pullRequest/create`,
        JSON.stringify({ pull_request_id: pr_id, pull_request_name: pr_name, author_id }),
        hdrs
    );

    if (res.status === 201) data.prs.push(pr_id);
    else console.error(`Ошибка при prCreate ${pr_id}: ${res.status}`);
}

export function mergePR(data) {
    if (!data.prs || data.prs.length === 0) return console.error('Нет PR для merge');
    const pr_id = randomItem(data.prs);

    const res = http.post(`${BASE}/pullRequest/merge`,
        JSON.stringify({ pull_request_id: pr_id }), hdrs);
    if (res.status !== 200) {
        console.error(`Ошибка при mergePR ${pr_id}: ${res.status}`);
    }
}

export function getTeamInfo(data) {
    const team_name = randomItem(data.teams);
    const res = http.get(`${BASE}/team/get?team_name=${team_name}`);
    if (res.status !== 200) console.error(`Ошибка при getTeam ${team_name}: ${res.status}`);
}

export function getUserPRs(data) {
    const user_id = randomItem(data.users);
    const res = http.get(`${BASE}/users/getReview?user_id=${user_id}`);
    if (res.status !== 200) {
        console.error(`Ошибка при getUserPrs : ${res.status}`);
    }
}

export function setIsActive(data) {
    const user_id = randomItem(data.users);
    const is_active = Math.random() < 0.5;

    const res = http.post(`${BASE}/users/setIsActive`,
        JSON.stringify({ user_id, is_active }), hdrs);
    if (res.status !== 200) {
        console.error(`Ошибка при setIsActive : ${res.status}`);
    }
}
