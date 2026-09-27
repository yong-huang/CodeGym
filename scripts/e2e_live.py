#!/usr/bin/env python3
"""CodeGym live end-to-end smoke test through the frontend proxy (:5190)."""
import json
import sys
import time
import urllib.request

BASE = "http://localhost:5190"
API = BASE + "/api/v1"
RUN = str(int(time.time()))
failures = []

# Bypass the macOS system proxy so localhost requests stay direct.
_NO_PROXY = urllib.request.build_opener(urllib.request.ProxyHandler({}))


def check(name, cond, detail=""):
    mark = "PASS" if cond else "FAIL"
    print(f"[{mark}] {name}" + (f" — {detail}" if detail and not cond else ""))
    if not cond:
        failures.append(name)


def get(url, headers=None):
    req = urllib.request.Request(url, headers=headers or {})
    with _NO_PROXY.open(req, timeout=15) as r:
        return r.status, r.read()


def post(url, payload, headers=None):
    data = json.dumps(payload).encode()
    h = {"Content-Type": "application/json", **(headers or {})}
    req = urllib.request.Request(url, data=data, headers=h, method="POST")
    with _NO_PROXY.open(req, timeout=15) as r:
        return r.status, json.loads(r.read())


def put(url, payload, headers=None):
    data = json.dumps(payload).encode()
    h = {"Content-Type": "application/json", **(headers or {})}
    req = urllib.request.Request(url, data=data, headers=h, method="PUT")
    with _NO_PROXY.open(req, timeout=15) as r:
        return r.status, json.loads(r.read())


def delete(url, headers=None):
    req = urllib.request.Request(url, headers=headers or {}, method="DELETE")
    with _NO_PROXY.open(req, timeout=15) as r:
        return r.status, r.read()


def status_of(url, headers=None):
    """Returns the HTTP status without raising on 4xx."""
    req = urllib.request.Request(url, headers=headers or {})
    try:
        with _NO_PROXY.open(req, timeout=15) as r:
            return r.status
    except urllib.error.HTTPError as e:
        return e.code


def bearer(token):
    return {"Authorization": f"Bearer {token}"}


def main():
    # 1. frontend shell
    st, html = get(BASE + "/")
    check("index.html serves", st == 200 and "<div id=\"root\">" in html.decode())

    for mod in ["src/main.tsx", "src/App.tsx", "src/pages/HomePage.tsx",
                "src/pages/TrainingPage.tsx", "src/pages/AdminPage.tsx",
                "src/components/CodeEditor.tsx"]:
        st, _ = get(f"{BASE}/{mod}")
        check(f"module transforms {mod}", st == 200)

    # 2. exercise list via proxy
    st, body = get(API + "/exercises")
    exs = json.loads(body)
    check("exercise list via proxy", st == 200 and len(exs) >= 7, f"{len(exs)} items")

    # 3. per-exercise variant across all weight levels
    exercises = {
        "py-lc-001": 3, "py-palindrome-002": 4, "py-factorial-bug-003": 1,
        "py-fizzbuzz-004": 4, "kubectl-getpods-005": 1, "k8s-nginx-pod-006": 3,
        "js-even-squares-007": 2,
        "go-even-squares-101": 3, "go-palindrome-102": 4, "go-factorial-bug-103": 1,
        "go-fizzbuzz-104": 4, "go-fib-105": 3, "go-concurrent-106": 4,
        "go-wordcount-107": 3, "go-map-missing-108": 1,
    }
    for ex_id, points in exercises.items():
        for lvl in ["beginner", "easy", "medium", "hard", "expert"]:
            st, body = get(f"{API}/exercises/{ex_id}/variant?level={lvl}")
            v = json.loads(body)
            ok = st == 200 and v["exercise_id"] == ex_id and v["variant_id"]
            ok = ok and "____FILL_" in v["marked_code"]
            ok = ok and len(v["fill_points"]) == v["marked_code"].count("____FILL_")
            ok = ok and all("original_text" not in p for p in v["fill_points"])
            check(f"variant {ex_id}@{lvl}", ok, f"{len(v['fill_points'])} fills")

    # 4. full training loop: fill exercise correct & wrong
    _, good = post(API + "/execute", {
        "exercise_id": "py-lc-001", "weight_level": "medium",
        "fill_answers": {"FILL_expr": "x**2", "FILL_cond": "x % 2 == 0", "FILL_iter": "numbers"},
    })
    check("fill: all correct -> S/100", good["is_correct"] and good["score"] == 100 and good["grade"] == "S",
          f"score={good['score']} grade={good['grade']}")
    check("fill: real execution output", good["output"] == "[4, 16, 36, 64, 100]", repr(good["output"]))

    _, wrong = post(API + "/execute", {
        "exercise_id": "py-lc-001", "weight_level": "medium",
        "fill_answers": {"FILL_expr": "x*2", "FILL_cond": "x % 2 == 0", "FILL_iter": "numbers"},
    })
    check("fill: wrong expr fails", not wrong["is_correct"] and wrong["score"] < 100,
          f"score={wrong['score']}")

    # 5. hint penalty
    _, hinted = post(API + "/execute", {
        "exercise_id": "kubectl-getpods-005", "weight_level": "easy",
        "fill_answers": {"FILL_flag": "-A"}, "hint_usage": {"flag": 4},
    })
    check("hint L4 penalty (-30)", hinted["score"] == 70, f"score={hinted['score']}")

    # 6. bugfix flow
    fx = "def factorial(n):\n    result = 1\n    for i in range(1, n + 1):\n        result *= i\n    return result"
    _, fixed = post(API + "/execute", {
        "exercise_id": "py-factorial-bug-003", "weight_level": "hard", "user_code": fx,
    })
    check("bugfix: correct fix -> S", fixed["is_correct"] and fixed["score"] == 100,
          f"score={fixed['score']} output={fixed['output']!r}")

    _, broken = post(API + "/execute", {
        "exercise_id": "py-factorial-bug-003", "weight_level": "hard",
        "user_code": "def factorial(n):\n    result = 1\n    for i in range(n):\n        result *= i\n    return result",
    })
    check("bugfix: broken fix -> D/0", not broken["is_correct"] and broken["score"] == 0,
          f"score={broken['score']}")

    # 7. config & command
    _, cfg = post(API + "/execute", {
        "exercise_id": "k8s-nginx-pod-006", "weight_level": "medium",
        "fill_answers": {"FILL_image": "nginx:1.25", "FILL_port": "80", "FILL_name": "nginx-pod"},
    })
    check("config: correct -> S", cfg["is_correct"] and cfg["score"] == 100, f"score={cfg['score']}")

    # Kubernetes 基础题库（50 题）
    _, k8s = get(API + "/exercises?tech_stack=kubernetes")
    k8s_list = json.loads(k8s)
    n_basics = sum(1 for e in k8s_list if e["id"].startswith("k8s-basics-"))
    check("k8s basic bank has 50 questions", n_basics == 50, f"{n_basics} found")
    _, k8cmd = post(API + "/execute", {
        "exercise_id": "k8s-basics-001", "weight_level": "beginner",
        "fill_answers": {"FILL_flag": "--all-namespaces"},
    })
    check("k8s command: acceptable alias graded", k8cmd["is_correct"] and k8cmd["score"] == 100,
          f"score={k8cmd['score']}")
    _, k8cfg = post(API + "/execute", {
        "exercise_id": "k8s-basics-051", "weight_level": "medium",
        "fill_answers": {"FILL_cpu": '"500m"', "FILL_mem": '"128Mi"'},
    })
    check("k8s config: yaml answer graded", k8cfg["is_correct"] and k8cfg["score"] == 100,
          f"score={k8cfg['score']}")

    # Go library: real `go run` execution
    _, goex = post(API + "/execute", {
        "exercise_id": "go-even-squares-101", "weight_level": "medium",
        "fill_answers": {"FILL_cond": "n%2 == 0", "FILL_square": "n*n", "FILL_append": "append"},
    })
    check("go: correct fills run & pass", goex["is_correct"] and goex["score"] == 100 and goex["output"] == "[4 16 36 64]",
          f"score={goex['score']} output={goex['output']!r}")
    gofix = ("package main\n\nimport \"fmt\"\n\nfunc factorial(n int) int {\n"
             "\tresult := 1\n\tfor i := 1; i <= n; i++ {\n\t\tresult *= i\n\t}\n\treturn result\n}\n\n"
             "func main() {\n\tfmt.Println(factorial(5))\n\tfmt.Println(factorial(3))\n}")
    _, gofx = post(API + "/execute", {
        "exercise_id": "go-factorial-bug-103", "weight_level": "hard", "user_code": gofix,
    })
    check("go: bugfix correct -> S", gofx["is_correct"] and gofx["score"] == 100,
          f"score={gofx['score']} output={gofx['output']!r}")

    # export / import round trip
    _, body = get(API + "/admin/export")
    exported = json.loads(body)
    check("export returns full library", len(exported) >= 15, f"{len(exported)} exercises")

    _, cmd = post(API + "/execute", {
        "exercise_id": "kubectl-getpods-005", "weight_level": "easy",
        "fill_answers": {"FILL_flag": "--all-namespaces"},
    })
    check("command: equivalent flag accepted", cmd["is_correct"] and cmd["score"] == 100,
          f"score={cmd['score']}")

    # browser sends bare point ids (no FILL_ prefix) -> must still grade correctly
    _, bare = post(API + "/execute", {
        "exercise_id": "py-palindrome-002", "weight_level": "medium",
        "fill_answers": {"return_expr": "s[::-1]", "return_type": "-> bool",
                         "compare": "s == s[::-1]"},
    })
    ok = bare["is_correct"] and bare["score"] == 100 and bare["grade"] == "S"
    ok = ok and not bare.get("weak_analysis")
    check("browser bare-key answers graded", ok, f"score={bare['score']} grade={bare['grade']}")

    # 8. 用户体系：注册 / 登录 / me / 我的题库（每个 run 独立账号）
    _, reg = post(API + "/auth/register",
                  {"username": f"e2elive{RUN}", "password": "secret123", "display_name": "冒烟账号"})
    token, live_user = reg["token"], reg["user"]
    check("register via proxy", bool(token) and live_user["username"].startswith("e2elive"))
    check("anonymous me rejected", status_of(API + "/auth/me") == 401)
    st, body = get(API + "/auth/me", bearer(token))
    check("me with token", st == 200 and json.loads(body)["user"]["id"] == live_user["id"])

    st, body = get(API + "/me/banks", bearer(token))
    banks = json.loads(body)
    check("bank picker available", st == 200 and len(banks["available"]) >= 50,
          f"{len(banks['available'])} banks")
    st, saved = put(API + "/me/banks", {"selected": ["testing::pytest", "bogus::x"]}, bearer(token))
    check("bank selection saved", st == 200 and saved["selected"] == ["testing::pytest"],
          repr(saved))
    st, body = get(API + "/auth/me", bearer(token))
    check("bank selection persisted", json.loads(body)["user"]["selected_banks"] == ["testing::pytest"])

    # progress + weak points for this account
    user = bearer(token)
    post(API + "/execute", {"exercise_id": "py-lc-001", "weight_level": "easy",
                            "fill_answers": {"FILL_expr": "x*2", "FILL_iter": "numbers"}}, user)
    post(API + "/execute", {"exercise_id": "py-lc-001", "weight_level": "easy",
                            "fill_answers": {"FILL_expr": "x**2", "FILL_iter": "numbers"}}, user)
    st, body = get(API + "/progress", user)
    prog = json.loads(body)
    p = prog["progress"]
    check("live progress tracked", p["total_attempts"] == 2 and p["correct_attempts"] == 1,
          f"attempts={p['total_attempts']}/{p['correct_attempts']}")
    check("live weak points recorded", p["weak_points"].get("py-lc-001") == 1,
          repr(p["weak_points"]))
    check("recent attempts populated", len(prog["recent_attempts"]) == 2)

    st, body = get(API + "/weak-points", user)
    weak = json.loads(body)
    check("weak-points endpoint", any(w["exercise_id"] == "py-lc-001" for w in weak))

    # 用户隔离：另一个账号看不到上面的进度
    _, reg2 = post(API + "/auth/register", {"username": f"e2eliveb{RUN}", "password": "secret123"})
    st, body = get(API + "/progress", bearer(reg2["token"]))
    other = json.loads(body)["progress"]
    check("user data isolated", other["total_attempts"] == 0, repr(other["total_attempts"]))

    # 教练端用户列表能看到新账号；跑完清理
    st, body = get(API + "/admin/users")
    users = json.loads(body)
    check("admin user list", st == 200 and any(u["id"] == live_user["id"] for u in users))
    for u in (live_user, reg2["user"]):
        delete(API + f"/admin/users/{u['id']}")

    # 9. coach/admin endpoints via proxy
    ex_id = f"e2e-proxy-{RUN}"
    sample = {
        "id": ex_id, "title": "代理测试", "description": "proxy smoke",
        "exercise_type": "fill", "tech_stack": "python", "weight": "easy",
        "category": "e2e", "full_code": "print(1)", "expected_output": "1",
        "fill_points": {"points": [{"id": "p1", "original_text": "1", "location": {"span_start": 6, "span_end": 7},
                                     "category": "expression", "min_weight": "easy", "weight_score": 1,
                                     "fill_type": "text"}], "min_fills": 1, "max_fills": 1},
        "enabled": True, "time_limit_seconds": 30,
    }
    _, created = post(API + "/admin/exercises", sample)
    check("admin create", created["id"] == ex_id)
    st, body = get(API + "/admin/stats")
    stats = json.loads(body)
    check("admin stats", stats["exercise_count"] >= 8 and stats["attempt_count"] >= 1,
          repr(stats))

    # cleanup: remove the exercise this run created so the bank stays clean
    req = urllib.request.Request(f"{API}/admin/exercises/{ex_id}", method="DELETE")
    _NO_PROXY.open(req, timeout=10).close()

    # favorites: toggle on/off for a dedicated live account
    _, fav_reg = post(API + "/auth/register", {"username": f"e2efav{RUN}", "password": "secret123"})
    fuser = bearer(fav_reg["token"])
    _, sky = post(API + "/favorites", {"exercise_id": "k8s-basics-001"}, fuser)
    check("favorite toggle on", sky.get("favorited") is True, repr(sky))
    _, body = get(API + "/favorites", fuser)
    favs = json.loads(body)
    check("favorites list for user", len(favs) == 1 and favs[0]["exercise_id"] == "k8s-basics-001",
          repr(favs))
    _, sky = post(API + "/favorites", {"exercise_id": "k8s-basics-001"}, fuser)
    check("favorite toggle off", sky.get("favorited") is False, repr(sky))
    delete(API + f"/admin/users/{fav_reg['user']['id']}")

    print()
    if failures:
        print(f"✗ {len(failures)} check(s) failed")
        for f in failures:
            print("   -", f)
        sys.exit(1)
    print("✓ ALL LIVE CHECKS PASSED")


if __name__ == "__main__":
    main()