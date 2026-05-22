import "../styles/tests.css";
import "../styles/confirm-modal.css";
import "../styles/tests-share-modal.css";

import LogoutButton from "../components/LogoutButton.jsx";
import { useState, useEffect, useRef } from "react";
import { useNavigate } from "react-router-dom";
import EditIcon from "../assets/edit.svg?react";
import ShareIcon from "../assets/share.svg?react";
import StatisticsIcon from "../assets/statistics.svg?react";
import CloseIcon from "../assets/close.svg?react";
import DeleteIcon from "../assets/delete.svg?react";
import CopyIcon from "../assets/copy_sub.svg?react";
import { testsAPI } from "../services/api.js";
import BackIcon from "../assets/back.svg?react";
import TaskIcon from "../assets/task.svg?react";
import EventIcon from "../assets/event.svg?react";
import CandidatesIcon from "../assets/Candidates.svg?react";
export default function Tests() {
    const [statsTest, setStatsTest] = useState(null);
    const navigate = useNavigate();


    const [tests, setTests] = useState([]);
    const [openMenuId, setOpenMenuId] = useState(null);
    const menuRefs = useRef({});
    const [shareModalOpen, setShareModalOpen] = useState(false);
    const [shareLink, setShareLink] = useState("");
    const [confirmModalOpen, setConfirmModalOpen] = useState(false);
    const [testToDelete, setTestToDelete] = useState(null);

    useEffect(() => {
        const fetchTests = async () => {
            try {
                const token = localStorage.getItem("token");
                if (!token) {
                    navigate("/login");
                    return;
                }

                const response = await testsAPI.getTests();
                const data = response.data;
                console.log("РџРѕР»СѓС‡РµРЅРЅС‹Рµ С‚РµСЃС‚С‹:", data);

                let testsArray = [];
                if (Array.isArray(data)) {
                    testsArray = data;
                } else if (data.tests && Array.isArray(data.tests)) {
                    testsArray = data.tests;
                } else if (data.data && Array.isArray(data.data)) {
                    testsArray = data.data;
                } else {
                    console.error("РќРµРёР·РІРµСЃС‚РЅР°СЏ СЃС‚СЂСѓРєС‚СѓСЂР° РѕС‚РІРµС‚Р°:", data);
                }

                const normalizedTests = testsArray.map(test => ({
                    ...test,
                    id: test.test_id || test.id,
                }));

                setTests(normalizedTests);
                console.log('ids:', normalizedTests.map(t => t.id));

            } catch (error) {
                console.error("РћС€РёР±РєР°:", error);
                alert("РќРµ СѓРґР°Р»РѕСЃСЊ Р·Р°РіСЂСѓР·РёС‚СЊ С‚РµСЃС‚С‹");
            }
        };

        fetchTests();
    }, [navigate]);

    const toggleMenu = (id, e) => {
        if (e) e.stopPropagation();
        setOpenMenuId(openMenuId === id ? null : id);
    };

    useEffect(() => {
        const handleClickOutside = (e) => {
            let clickedInsideMenu = false;

            Object.values(menuRefs.current).forEach(ref => {
                if (ref && ref.contains(e.target)) {
                    clickedInsideMenu = true;
                }
            });

            if (!clickedInsideMenu) {
                setOpenMenuId(null);
            }
        };

        document.addEventListener("mousedown", handleClickOutside);
        return () => document.removeEventListener("mousedown", handleClickOutside);
    }, []);

    const editTest = (test) => {
        console.log("РўРµСЃС‚ РґР»СЏ СЂРµРґР°РєС‚РёСЂРѕРІР°РЅРёСЏ:", test);
        navigate("/create", {
            state: { editing: true, test: test, deleteOnSave: true },
        });
        setOpenMenuId(null);
    };




    const deleteTest = async (id) => {
        try {
            const token = localStorage.getItem("token");
            if (!token) {
                alert("РўСЂРµР±СѓРµС‚СЃСЏ Р°РІС‚РѕСЂРёР·Р°С†РёСЏ");
                navigate("/login");
                return;
            }

            console.log("РЈРґР°Р»РµРЅРёРµ С‚РµСЃС‚Р° СЃ ID:", id);
            await testsAPI.deleteTest(id);

            const updatedTests = tests.filter(test => {
                const testId = test.id;
                return testId !== id;
            });

            setTests(updatedTests);
            setOpenMenuId(null);
            setConfirmModalOpen(false);

        } catch (error) {
            console.error("РћС€РёР±РєР° РїСЂРё СѓРґР°Р»РµРЅРёРё С‚РµСЃС‚Р°:", error);
            alert("РќРµ СѓРґР°Р»РѕСЃСЊ СѓРґР°Р»РёС‚СЊ С‚РµСЃС‚ РЅР° СЃРµСЂРІРµСЂРµ. РџСЂРѕРІРµСЂСЊС‚Рµ РєРѕРЅСЃРѕР»СЊ РґР»СЏ РґРµС‚Р°Р»РµР№.");
        }
    };


    const openDeleteConfirm = (test) => {
        setTestToDelete(test);
        setConfirmModalOpen(true);
        setOpenMenuId(null);
    };


    const closeDeleteConfirm = () => {
        setConfirmModalOpen(false);
        setTestToDelete(null);
    };;

    const shareTest = async (test) => {
        try {

            const link = `${window.location.origin}/test/${test.test_link}`;
            setShareLink(link);
            setShareModalOpen(true);
        } catch (error) {
            console.error("РћС€РёР±РєР° РїСЂРё РїРѕРґРіРѕС‚РѕРІРєРµ СЃСЃС‹Р»РєРё:", error);
            alert("РќРµ СѓРґР°Р»РѕСЃСЊ РїРѕРґРіРѕС‚РѕРІРёС‚СЊ СЃСЃС‹Р»РєСѓ");
        }
        setOpenMenuId(null);
    };





    const closeTest = async (id) => {

    };

    const viewStatistics = (test) => {
        navigate(`/statistics/${test.id}`);
        setOpenMenuId(null);
    };



    return (
        <div className="tests-page">
            <>
                <LogoutButton />
            </>
            <div className="tests-wrapper">
                <div className="tests-left">
                    {/* РќР°РІРёРіР°С†РёРѕРЅРЅС‹Рµ РІРєР»Р°РґРєРё */}
                    <div className="tests-tabs">
                        <button
                            className="tab-btn tab-btn-active"
                            onClick={() => navigate("/tests")}
                        >
                            <TaskIcon />
                            РўРµСЃС‚РѕРІС‹Рµ Р·Р°РґР°РЅРёСЏ
                        </button>
                        <button
                            className="tab-btn"
                            onClick={() => navigate("/events")}
                        >
                            <EventIcon />
                            РњРµСЂРѕРїСЂРёСЏС‚РёСЏ
                        </button>
                        <button
                            className="tab-btn"
                            onClick={() => navigate("/candidates")}
                        >
                            <CandidatesIcon />
                            РљР°РЅРґРёРґР°С‚С‹
                        </button>
                    </div>
                    {/* <div className="tests-line"></div> */}


                    {tests.length === 0 ? (
                        <div className="no-tests">
                            РџРѕРєР° РЅРµС‚ С‚РµСЃС‚РѕРІ. РЎРѕР·РґР°Р№С‚Рµ РїРµСЂРІС‹Р№ С‚РµСЃС‚ в†’
                        </div>
                    ) : (
                        <div className="tests-grid">
                            {tests.map((test) => {

                                const testId = test.id;
                                const testTitle = test.Title || test.title;
                                const isActive = test.IsActive !== false;

                                return (
                                    <div key={testId} className="test-card"
                                         style={{
                                             zIndex: openMenuId === testId ? 100 : 1,
                                             opacity: isActive ? 1 : 0.6
                                         }}
                                    >
                                        <div
                                            className="test-menu-container"
                                            ref={el => menuRefs.current[testId] = el}
                                        >
                                            <button
                                                className="dots-btn"
                                                onClick={(e) => toggleMenu(testId, e)}
                                            >
                                                в‹®
                                            </button>

                                            {openMenuId === testId && (
                                                <div className="dropdown-menu">
                                                    <button className="menu-item" onClick={() => editTest(test)}>
                                                        <EditIcon className="menu-icon" />
                                                        <span>Р РµРґР°РєС‚РёСЂРѕРІР°С‚СЊ</span>
                                                    </button>
                                                    <button className="menu-item share" onClick={() => shareTest(test)}>
                                                        <ShareIcon className="menu-icon" />
                                                        <span>РџРѕРґРµР»РёС‚СЊСЃСЏ</span>
                                                    </button>
                                                    <button className="menu-item" onClick={() => viewStatistics(test)}>
                                                        <StatisticsIcon className="menu-icon" />
                                                        <span>РЎС‚Р°С‚РёСЃС‚РёРєР°</span>
                                                    </button>

                                                    <button className="menu-item" onClick={() => openDeleteConfirm(test)}>
                                                        <DeleteIcon className="menu-icon" />
                                                        <span>РЈРґР°Р»РёС‚СЊ С‚РµСЃС‚</span>
                                                    </button>
                                                </div>
                                            )}
                                        </div>
                                        <span className="test-titles">
                                            {testTitle && testTitle.length > 15
                                                ? `${testTitle.substring(0, 15)}...`
                                                : testTitle || "Р‘РµР· РЅР°Р·РІР°РЅРёСЏ"
                                            }
                                        </span>
                                        {!isActive && (
                                            <div className="test-status">Р—РђРљР Р«Рў</div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>

                <div className="tests-right">
                    <button className="create-test-btn" onClick={() => navigate("/create")}>
                        РЎРѕР·РґР°С‚СЊ С‚РµСЃС‚
                    </button>
                </div>
            </div>
            {shareModalOpen && (
                <div className="share-modal-overlay" onClick={() => setShareModalOpen(false)}>
                    <div
                        className="share-modal"
                        onClick={(e) => e.stopPropagation()}
                    >
                        <h3 className="share-modal-title">РџРѕРґРµР»РёС‚СЊСЃСЏ СЃСЃС‹Р»РєРѕР№</h3>

                        <div className="share-modal-body">
                            <input
                                type="text"
                                className="share-modal-input"
                                value={shareLink}
                                readOnly
                            />
                            <button
                                className="share-modal-copy-btn"
                                onClick={async () => {
                                    try {
                                        await navigator.clipboard.writeText(shareLink);
                                    } catch (e) {
                                        console.error("РћС€РёР±РєР° РєРѕРїРёСЂРѕРІР°РЅРёСЏ:", e);
                                        alert("РќРµ СѓРґР°Р»РѕСЃСЊ СЃРєРѕРїРёСЂРѕРІР°С‚СЊ СЃСЃС‹Р»РєСѓ");
                                    }
                                }}
                            >
                                <CopyIcon className="share-modal-copy-icon" />
                            </button>

                        </div>
                    </div>
                </div>
            )}
            {confirmModalOpen && testToDelete && (
                <div className="confirm-modal-overlay" onClick={closeDeleteConfirm}>
                    <div
                        className="confirm-modal"
                        onClick={(e) => e.stopPropagation()}
                    >
                        <h3 className="confirm-modal-title">РЈРґР°Р»РёС‚СЊ С‚РµСЃС‚</h3>
                        <p className="confirm-modal-message">
                            Р’С‹ СѓРІРµСЂРµРЅС‹, С‡С‚Рѕ С…РѕС‚РёС‚Рµ СѓРґР°Р»РёС‚СЊ С‚РµСЃС‚
                            <strong> "{testToDelete.Title || testToDelete.title || "Р‘РµР· РЅР°Р·РІР°РЅРёСЏ"}"</strong>?
                            <br />
                        </p>
                        <div className="confirm-modal-buttons">
                            <button
                                className="confirm-modal-btn confirm-modal-btn-cancel"
                                onClick={closeDeleteConfirm}
                            >
                                РћС‚РјРµРЅР°
                            </button>
                            <button
                                className="confirm-modal-btn confirm-modal-btn-delete"
                                onClick={() => deleteTest(testToDelete.id)}
                            >
                                РЈРґР°Р»РёС‚СЊ
                            </button>
                        </div>
                    </div>
                </div>
            )}
            {statsTest && (
                <StatisticsTest
                    testId={statsTest.id}
                    onClose={() => setStatsTest(null)}
                />
            )}

        </div>
    );
}
