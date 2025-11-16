import http from 'k6/http';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const hdrs = { headers: { 'Content-Type': 'application/json' } };

export const options = {
    scenarios: {
        reassign_reviewer: {
            executor: 'constant-arrival-rate',
            rate: 100,
            timeUnit: '1s',
            duration: '2m',
            preAllocatedVUs: 1,
            exec: 'reassignReviewer',
        },
    },
};

export function setup() {
    const teams = [];
    const users = [];
    const teamMembers = {};
    const userTeam = {};


   for (let i = 0; i < 15; i++) {
        const team_name = `Team${i}`;
        const members = [];
        const userIds = [];

        for (let j = 0; j < 20; j++) {
            const user_id = `u${i}_${j}`;
            members.push({ user_id, username: `User${i}_${j}`, is_active: true });
            userIds.push(user_id);
            users.push(user_id);
            userTeam[user_id] = team_name;
        }

        teamMembers[team_name] = userIds;

        const res = http.post(`${BASE}/team/add`, JSON.stringify({ team_name, members }), hdrs);
        if (res.status !== 201 && res.status !== 200) {
            console.error(`Ошибка при создании команды ${team_name}: ${res.status}, body: ${res.body}`);
        }

        teams.push(team_name);
    }


    const prs = {}; // pr_id -> { pr_id, pr_author, pr_status, pr_reviewers }
    users.forEach((user_id, idx) => {
        const pr_id = `pr-${idx}`;
        const pr_name = `Feature${idx}`;

        const res = http.post(
            `${BASE}/pullRequest/create`,
            JSON.stringify({ pull_request_id: pr_id, pull_request_name: pr_name, author_id: user_id }),
            hdrs
        );

        if (res.status === 201) {
            const prJson = res.json().pr;
            prs[pr_id] = {
                pr_id,
                pr_author: user_id,
                pr_status: 'open',
                pr_reviewers: prJson.assigned_reviewers || [],
            };
        } else {
            console.error(`Ошибка при создании PR ${pr_id}: ${res.status}`);
        }

    });

    return { teams, users, prs, userTeam, teamMembers };
}

export function reassignReviewer(data) {
    if (!data.prs || Object.keys(data.prs).length === 0) return;

    const prsToReassign = [];
    const prKeys = Object.keys(data.prs);

    for (let i = 0; i < 10; i++) {
        const pr_id = randomItem(prKeys);
        const pr = data.prs[pr_id];
        if (pr.pr_reviewers && pr.pr_reviewers.length > 0) {
            const old_user_id = randomItem(pr.pr_reviewers);
            prsToReassign.push({ pr_id, old_user_id, pr });
        }
    }

    const requests = prsToReassign.map(({ pr_id, old_user_id }) => {
        return [
            'POST',
            `${BASE}/pullRequest/reassign`,
            JSON.stringify({ pull_request_id: pr_id, old_user_id }),
            { headers: { 'Content-Type': 'application/json' } },
        ];
    });

    const responses = http.batch(requests);

    responses.forEach((res, idx) => {
        const { pr_id, old_user_id, pr } = prsToReassign[idx];
        if (res.status === 200) {
            const body = res.json();
            const new_id = body.replaced_by;
            const idxReviewer = pr.pr_reviewers.indexOf(old_user_id);
            if (idxReviewer >= 0) pr.pr_reviewers[idxReviewer] = new_id;
        } else {
            console.error(`Ошибка при переназначении PR ${pr_id}: ${res.status}`);
        }
    });
}