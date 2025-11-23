import http from 'k6/http';
import { check, sleep } from 'k6';

// Настройки нагрузки
export let options = {
    vus: 5,       // виртуальные пользователи
    iterations: 50 // всего запросов на каждый эндпоинт
};

// Функция генерации рандомного ID
function randID() {
    return Math.floor(Math.random() * 1000000);
}

// Основная функция
export default function () {

    // /team/add
    let teamName = `team-${randID()}`;
    let user1 = `u${randID()}`;
    let user2 = `u${randID()}`;
    let user3 = `u${randID()}`;

    let teamPayload = JSON.stringify({
        team_name: teamName,
        members: [
            { user_id: user1, username: "Alice", is_active: false },
            { user_id: user2, username: "Bob", is_active: true },
            { user_id: user3, username: "Danils", is_active: true }
        ]
    });

    let res = http.post('http://localhost:8080/team/add', teamPayload, {
        headers: { 'Content-Type': 'application/json' }
    });
    check(res, { 'team added': (r) => r.status === 201 });

    sleep(0.1);

    // /users/setIsActive
    let setActivePayload = JSON.stringify({ user_id: user1, is_active: true });
    res = http.post('http://localhost:8080/users/setIsActive', setActivePayload, {
        headers: { 'Content-Type': 'application/json' }
    });
    check(res, { 'user deactivated': (r) => r.status === 200 });

    sleep(0.1);

    // /pullRequest/create
    let prID = `pr-${randID()}`;
    let prPayload = JSON.stringify({
        pull_request_id: prID,
        pull_request_name: `PR ${randID()}`,
        author_id: user1
    });
    res = http.post('http://localhost:8080/pullRequest/create', prPayload, {
        headers: { 'Content-Type': 'application/json' }
    });
    check(res, { 'PR created': (r) => r.status === 201 });

    sleep(0.1);

    // /pullRequest/merge
    let mergePayload = JSON.stringify({ pull_request_id: prID });
    res = http.post('http://localhost:8080/pullRequest/merge', mergePayload, {
        headers: { 'Content-Type': 'application/json' }
    });
    check(res, { 'PR merged': (r) => r.status === 200 });

    sleep(0.1);

    // /team/get
    res = http.get(`http://localhost:8080/team/get?team_name=${teamName}`);
    check(res, { 'team fetched': (r) => r.status === 200 });

    sleep(0.1);

    // /users/getReview
    res = http.get(`http://localhost:8080/users/getReview?user_id=${user1}`);
    check(res, { 'user reviews fetched': (r) => r.status === 200 });
}