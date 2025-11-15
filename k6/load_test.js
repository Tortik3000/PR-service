import http from 'k6/http';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE = __ENV.BASE_URL || `http://localhost:8080`;
const hdrs = { headers: { 'Content-Type': 'application/json' } };

function randomNum() {
    return Math.floor(Math.random() * 1_000_000);
}

export const options = {
    scenarios: {
        add_team: {
            executor: 'constant-arrival-rate',
            rate: 10,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'addTeam',
        },
        create_pr: {
            executor: 'constant-arrival-rate',
            rate: 10,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'createPR',
        },
        merge_pr: {
            executor: 'constant-arrival-rate',
            rate: 10,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'mergePR',
        },
        reassign_reviewer: {
            executor: 'constant-arrival-rate',
            rate: 10,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'reassignReviewer',
        },
        get_team_info: {
            executor: 'constant-arrival-rate',
             rate: 10,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'getTeamInfo',
        },
        get_user_prs: {
            executor: 'constant-arrival-rate',
            rate: 10,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 10,
            exec: 'getUserPRs',
        },
    },
};

export function setup() {
    const teams = [];
    const users = [];
    const teamMembers = {};
    const userTeam = {};
    const prAuthor = {};

    // Создаём команды
    for (let i = 0; i < 10; i++) {
        const team_name = `Team${i}`;
        const members = [];
        const userIds = [];

        for (let j = 0; j < 7; j++) {
            const user_id = `u${i}_${j}`;
            members.push({ user_id, username: `User${i}_${j}`, is_active: true });
            userIds.push(user_id);
            users.push(user_id);
            userTeam[user_id] = team_name;
        }

        teamMembers[team_name] = userIds;
        http.post(`${BASE}/team/add`, JSON.stringify({ team_name, members }), hdrs);
        teams.push(team_name);
    }

    // Создаём PR
    const prs = [];
    users.forEach((user_id, idx) => {
        const pr_id = `pr-${idx}`;
        const pr_name = `Feature${idx}`;
        const res = http.post(
            `${BASE}/pullRequest/create`,
            JSON.stringify({ pull_request_id: pr_id, pull_request_name: pr_name, author_id: user_id }),
            hdrs
        );
        if (res.status === 201) {
            prs.push(pr_id);
            prAuthor[pr_id] = user_id;
        }
    });

    return { teams, users, prs, prAuthor, userTeam, teamMembers };
}


export function addTeam() {
    const team_name = `Team${randomNum()}`;
    const members = [
        { user_id: `u${randomNum()}`, username: `User${randomNum()}`, is_active: true },
        { user_id: `u${randomNum()}`, username: `User${randomNum()}`, is_active: true },
    ];
    http.post(`${BASE}/team/add`, JSON.stringify({ team_name, members }), hdrs);
}

export function getTeamInfo(data) {
    http.get(`${BASE}/team/get?team_name=${randomItem(data.teams)}`);
}

export function createPR(data) {
    const author_id = randomItem(data.users);
    const pr_id = `pr-${randomNum()}`;
    const pr_name = `Feature${randomNum()}`;
    http.post(`${BASE}/pullRequest/create`, JSON.stringify({ pull_request_id: pr_id, pull_request_name: pr_name, author_id }), hdrs);
    data.prs.push(pr_id);
}

export function mergePR(data) {
    if (data.prs.length === 0) return;
    const pr_id = randomItem(data.prs);
    http.post(`${BASE}/pullRequest/merge`, JSON.stringify({ pull_request_id: pr_id }), hdrs);
}

export function reassignReviewer(data) {
    if (data.prs.length === 0) return;

    const pr_id = randomItem(data.prs);
    const author_id = data.prAuthor[pr_id];
    if (!author_id) return;

    const team = data.userTeam[author_id];
    if (!team) return;

    const members = data.teamMembers[team];
    if (!members || members.length === 0) return;
    
    const old_user_id = randomItem(members);

    http.post(
        `${BASE}/pullRequest/reassign`,
        JSON.stringify({ pull_request_id: pr_id, old_user_id }),
        hdrs
    );
}

export function getUserPRs(data) {
    const user_id = randomItem(data.users);
    http.get(`${BASE}/users/getReview?user_id=${user_id}`);
}
